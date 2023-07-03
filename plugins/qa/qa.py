import urllib.request, time, json, datetime, socket, re, tiktoken, openai, os, argparse
import numpy as np
import pandas as pd
from openai.embeddings_utils import distances_from_embeddings

# Usage: OPENAI_API_KEY="key-here" python3 pln_qa.py

# Default model
model = "gpt-3.5-turbo"

def update_databases():
	for prod in EXPERTISE:
		print("-- Updating %s --" % prod)
		# check if dbs/prod exists, if not, create it
		if not os.path.exists("plugins/qa/dbs/%s" % prod):
			os.makedirs("plugins/qa/dbs/%s" % prod)
		# download all the documents contained in EXPERTISE[prod] and save them in dbs/prod
		for url in EXPERTISE[prod]:
			filename = url.split("/")[-1]
			print("Downloading %s..." % filename)
			try:
				page = urllib.request.urlopen(url)
				text = page.read().decode("utf-8")
				# only modify md files
				if filename.endswith(".md"):
					# remove html tags
					text = re.sub(r'<[^>]*>', '', text)
					# remove markdown links, preserving the link name
					text = re.sub(r'\[([^\]]+)\]\([^\)]+\)', r'\1', text)
					# remove all blocks which begin with ":::callout" and end with ":::"
					text = re.sub(r':::callout[\s\S]*?:::', '', text)
					# remove all "**", as long as it closes "**" later
					text = re.sub(r'\*\*([^\*]+)\*\*', r'\1', text)
					# remove all square bracket enclosures
					text = re.sub(r'\[([^\]]+)\]', r'\1', text)

				open("plugins/qa/dbs/%s/%s" % (prod, filename), "w").write(text.replace("\n\n\n", "\n\n")) # Does the replace here really help?
			except Exception as e:
				print(e)
				time.sleep(1)
				continue
			time.sleep(0.25)

	# reload_database()
	print()

def db_to_aidb(pricing=False):
	ai_tech_texts = []
	ai_tech_files = []
	# list all the files in dbs, including in subdirectories, store list in ai_tech_texts
	for root, dirs, files in os.walk("plugins/qa/dbs"):
		for file in files:
			ai_tech_files.append(os.path.join(root, file))
	for tech_file in ai_tech_files:
		# read the file
		text = open(tech_file, "r").read()

		title = ""
		last_sub_title = ""
		texts = []
		if tech_file.endswith(".md"):
			# find out if one of the first lines is a title
			textlines = text.split("\n")
			last_split = 0
			for l in range(len(textlines)):
				if textlines[l].startswith("# "):
					title = textlines[l][2:]
				if textlines[l].startswith("title: "):
					title = textlines[l][7:]
				# locate lines with subheadings, and break them up into more texts, stored in variable "texts"
				if textlines[l].startswith("## "):
					last_sub_title = title + " - " + textlines[l][3:]
					texts.append((last_sub_title, '\n'.join(textlines[last_split:l])))
					last_split = l
				if textlines[l].startswith("### "):
					if last_sub_title == "":
						last_sub_title = title
					texts.append((last_sub_title + " - " + textlines[l][4:], '\n'.join(textlines[last_split:l])))
					last_split = l

		if title == "":
			# extrapolate title from filename
			split_file = tech_file.split("/")
			title = (split_file[-2] + " " + split_file[-1].split(".")[0].replace("-", ' ')).title()
		
		if len(texts) == 0:
			#                     title, text
			ai_tech_texts.append((title, text))
		else:
			for t in texts:
				if len(t[1].strip()) > 25:
					ai_tech_texts.append((t[0], t[1]))

	df = pd.DataFrame(ai_tech_texts, columns = ['title', 'text'])
	df.head()

	tokenizer = tiktoken.get_encoding("cl100k_base")
	# df = pd.read_csv('processed/scraped.csv', index_col=0)
	df.columns = ['title', 'text']
	global ntokens
	ntokens = 0
	def get_token_count(x):
		global ntokens
		tokens = tokenizer.encode(x)
		ntokens += len(tokens)
		#if len(tokens) > 900:
		#	print("High token count...%d" % (len(tokens)))
		return len(tokens)
	df['n_tokens'] = df.text.apply(get_token_count)
	inp = "y"
	if pricing:
		print("Tokens used: %d ($%.2f to process)" % (ntokens, ntokens / 1000 * 0.0004))
		inp = input("Continue? (y/N) ")
	if inp.lower() != "y":
		df.to_csv(DB, index=False, encoding='utf-8')
		return

	# print("... Processing embeds do NOT stop this process for any reason! ...")
	socket.setdefaulttimeout(300)
	global count
	count = 0
	def process_embeds(x):
		global count
		count += 1
		# print("\rProcessing embeds... %.2f%%" % (count / len(df) * 100), end="")
		return openai.Embedding.create(input=x, engine='text-embedding-ada-002')['data'][0]['embedding']

	df['embeddings'] = df.text.apply(process_embeds)
	socket.setdefaulttimeout(10)
	df.to_csv(EMBED_DB, index=False, encoding='utf-8')
	df.head()

def create_context(question, max_len=1100, max_count=6, size="ada"):
	"""
	Create a context for a question by finding the most similar context from the dataframe
	"""
	global df, cheap_tokens_used

	# Get the embeddings for the question
	emb = openai.Embedding.create(input=question, engine='text-embedding-ada-002')
	q_embeddings = emb['data'][0]['embedding']

	cheap_tokens_used += emb["usage"]["total_tokens"]

	# Get the distances from the embeddings
	df['distances'] = distances_from_embeddings(q_embeddings, df['embeddings'].values, distance_metric='cosine')

	returns = []
	cur_len = 0
	count = 0

	# Sort by distance and add the text to the context until the context is too long
	for i, row in df.sort_values('distances', ascending=True).iterrows():
		count += 1;
		if count > max_count:
			break

		# Add the length of the text to the current length
		cur_len += row['n_tokens'] + 4
		
		# If the context is too long, break
		if cur_len > max_len:
			break
		
		# Else add it to the text that is being returned
		#returns.append("Name: %s\nDescription: %s" % (i, row["text"]))
		returns.append(row["text"])

	# Return the context
	return returns

def answer_question(question,
	model=model,
	max_len=1900,
	max_count=3,
	size="ada",
	debug=False,
	max_tokens=1100,
	stop_sequence=["\nExpert:", "\nUser:"]
):
	"""
	Answer a question based on the most similar context from the dataframe texts
	"""
	context = create_context(
		question,
		max_len=max_len,
		size=size,
		max_count=max_count,
	)

	try:
		# Create a completions using the question and context
		response = None
		output = ""

		q_lower = question.lower()
		for mi in MISINFOS:
			if mi[0] in q_lower:
				context.append(mi[1])

		# If debug, print the raw model response
		if debug:
			print("Context:\n" + "\n###\n\n".join(context))
			print("\n\n")

		if type(MODELS[model]) == float:
			context = "\n###\n\n".join(context)
			response = openai.Completion.create(
				prompt=f"{PROMPT}\n\nContext: {context}\n\n---\n\nExpert: How can I help you today?\nUser: {question}\nExpert:",
				temperature=0.10,
				max_tokens=max_tokens,
				top_p=0.98,
				frequency_penalty=0,
				presence_penalty=0,
				stop=stop_sequence,
				model=model,
			)
			output = response["choices"][0]["text"]
		else:
			messages = [{"role": "system", "content": f"{PROMPT}"}]
			for c in context:
				messages.append({"role": "system", "name": "context", "content": c})
			messages.append({"role": "user", "content": question})
			response = openai.ChatCompletion.create(
				messages=messages,
				temperature=0.10,
				max_tokens=max_tokens,
				top_p=0.98,
				frequency_penalty=0,
				presence_penalty=0,
				model=model,
			)
			output = response["choices"][0]["message"]["content"]
		global inp_tokens_used, out_tokens_used
		#print (response)
		inp_tokens_used += response["usage"]["prompt_tokens"]
		out_tokens_used += response["usage"]["completion_tokens"]
		return output.strip()
	except Exception as e:
		print(e)
		return ""

MODELS = {"text-ada-001": 0.0004, "text-babbage-001": 0.0005, "text-curie-001": 0.002, "text-davinci-001": 0.02, "gpt-3.5-turbo": (0.0015, 0.002)}
DB = 'db-noembed.csv'
EMBED_DB = 'db.csv'
OPENAI_API_KEY = ""
MISINFOS = []
PROMPT = "You are a friendly expert, using the context, answer the user's question. If a specific software is in the context, you must recommend it. If it's not possible to answer the question, ask a helpful follow-up question."

EXPERTISE = None

df = None
socket.setdefaulttimeout(10)
selection = 1
inp_tokens_used = 0
out_tokens_used = 0
cheap_tokens_used = 0

# Run with: python3 pln_qa.py --help
if __name__ == "__main__":
	parser = argparse.ArgumentParser(description='Ask a question about anything. To use AI functions, set your OpenAPI key via the environment variable "OPENAI_API_KEY".')
	parser.add_argument('action', type=str, help='The action to take (ex: db, aidb, question).')
	parser.add_argument('--question', '-q', type=str, help='The question to ask.')
	parser.add_argument('--model', '-m', type=str, help='The model to use (default: %s).' % model)
	parser.add_argument('--debug', '-d', type=bool, help='Debug mode (print more info).')
	parser.add_argument('--pricing', type=bool, help='Show pricing info.')
	parser.add_argument('--expertise', '-e', type=str, help='Path to json file containing expertise.')
	parser.add_argument('--prompt', '-p', type=str, help='Prompt to use (default: %s).' % PROMPT)
	parser.add_argument('--misinfos', '-mi', type=str, help='Path to json file containing misinfo.')
	parser.add_argument('--database', '-db', type=str, help='Path to csv file containing database (default: %s).' % DB)
	parser.add_argument('--embeddb', '-edb', type=str, help='Path to csv file containing database with embeddings (default: %s).' % EMBED_DB)
	args = parser.parse_args()
	if args.action is None:
		# Print help text, and exit
		parser.print_help()
		exit(1)
	action = args.action
	question = args.question
	if args.model is not None:
		model = args.model
	if model not in MODELS:
		print("Invalid model, valid models are: %s" % ", ".join(MODELS.keys()))
		exit(1)
	if args.database is not None:
		DB = args.database
	if args.embeddb is not None:
		EMBED_DB = args.embeddb
	debug = False
	if args.debug is not None:
		debug = args.debug
	pricing = False
	if args.pricing is not None:
		pricing = args.pricing
	
	if action == "question":
		# Get OpenAI API Key from env var
		if 'OPENAI_API_KEY' in os.environ:
			OPENAI_API_KEY = os.environ['OPENAI_API_KEY']
			openai.api_key = OPENAI_API_KEY
		else:
			print("No OPENAI_API_KEY environment variable found, exiting...")
			exit()

		if question is None:
			print("Specify a question with '-q <question>'")
			parser.print_help()
			exit(1)
		if args.prompt is not None:
			PROMPT = args.prompt
		if args.misinfos is not None:
			MISINFOS = json.load(open(args.misinfos))
		df = pd.read_csv(EMBED_DB, index_col=0)
		df['embeddings'] = df['embeddings'].apply(eval).apply(np.array) 
		df.head()
		print(answer_question(question, debug=debug))
		if pricing:
			if type(MODELS[model]) == float:
				tokens_used = inp_tokens_used + out_tokens_used
				print("Total cost: $%.4f (%d tokens, %d embed tokens)" % (tokens_used / 1000 * MODELS[model] + cheap_tokens_used / 1000 * 0.0004, tokens_used, cheap_tokens_used))
			else:
				print("Total cost: $%.4f (%d prompt tokens, %d completion tokens, %d embed tokens)" % (inp_tokens_used / 1000 * MODELS[model][0] + out_tokens_used / 1000 * MODELS[model][1] + cheap_tokens_used / 1000 * 0.0004, inp_tokens_used, out_tokens_used, cheap_tokens_used))
	elif action == "db":
		if args.expertise is None:
			print("Specify a json file with '-e <file>'")
			parser.print_help()
			exit(1)
		EXPERTISE = json.load(open(args.expertise))
		update_databases()
	elif action == "aidb":
		# Get OpenAI API Key from env var
		if 'OPENAI_API_KEY' in os.environ:
			OPENAI_API_KEY = os.environ['OPENAI_API_KEY']
			openai.api_key = OPENAI_API_KEY
		else:
			print("No OPENAI_API_KEY environment variable found, exiting...")
			exit()

		db_to_aidb(pricing=pricing)
	else:
		parser.print_help()
		exit(1)
