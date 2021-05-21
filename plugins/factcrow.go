// Copyright (c) 2020, The OneBot Contributors. All rights reserved.

package main

import (
	"github.com/TheDiscordian/onebot/onelib"
	"math/rand"
)

// Inspired by Fact Crow: https://gbatemp.net/threads/fact-crow-homebrew.376503/

const (
	// NAME is same as filename, minus extension
	NAME = "factcrow"
	// LONGNAME is what's presented to the user
	LONGNAME = "Fact Crow Plugin"
	// VERSION of the plugin
	VERSION = "v0.0.0"
)

var (
	factCrowAnswers = []string{"The critically endangered Kakapo bird has a strong, pleasant, musty odour which allows predators to easily locate it",
		"The tallest mountain in South Africa is Mount Kilimanjaro",
		"The answer you're searching for, to help explain why you exist, is as disappointing as the life you live",
		"The tomato is the national vegetable of New Jersey",
		"Your dreams will become your nightmares",
		"The next sleep you take could be your last",
		"You're the reason the grass is greener on the other side",
		"You spend your whole life stuck in the labyrinth, thinking about how you'll escape one day, and how awesome it will be, and imagining that future keeps you going, but you never do it. You just use the future to escape the present",
		"Your flesh was constructed upon the graves of hundreds of innocent beasts",
		"Every breath you take brings you one instant closer to your inevitable demise",
		"There is no God",
		"Your demise will go largely unnoticed",
		"The Declaration of Independance was adopted by the Continental Congress on July 4, 1776",
		"The world's largest rubber band ball is 9,032 pounds, and more than six-feet tall and more than 700,000 rubber bands",
		"Thousands of tiny lives die within you as you read this very sentence",
		"Greater men than you have, do, and always will exist in all of your pursuits",
		"The Mississippi River is the largest river system in the United States and the largest of North America",
		"The per capita GDP of france is $33,744",
		"Your birth was an accident regretted by all who were involved",
		"79 is the 22nd prime number",
		"Red, green, and blue are the primary colours of light",
		"I sit by your windowsill, awaiting your timely end",
		"Fear is the manifestation of your personal inadequacies",
		"True love awaits no one",
		"Tetris was created by Alexey Pazhitnov on June 6, 1984",
		"No afterlife awaits you",
		"Systematic attempts to evolve a system of perspective are usually considered to have begun around the 5th century B.C. in the art of Ancient Greece",
		"Your body and mind grow weaker every day as the beasts outside grow ever stronger",
		"The Dead Sea is one of the world's saltiest bodies of water, with 33.7% salinity",
		"Nothing can silence the voices",
		"My children hunger for your blood",
		"The atomic symbol for gold is 'Au'",
		"Nothing is permenant, everything you know will eventually come to an end",
		"Every action you take may result in your death",
		"Your family prays for your absence",
		"Granite is part of the felsic rock family",
		"Your capitalist war machine cannot feed itself indefinitly",
		"In the grand scheme of things, dust is no less significant than you",
		"A meteor is the visible streak of light that occurs when a meteoroid enters the Earth's atmosphere",
		"Life may exist on other planets, but it cares not of yours",
		"A fruit is a structure of a plant that contains its seeds",
		"There is no escape",
		"Oxygen gas constitutes 20.9% of the volume of air",
		"No medicine will cure the sickness that lives within you",
		"Criminals have a more positive effect on society than you",
		"Zeus is the King of the Gods in Greek Mythology",
		"Nothing can escape the cold embrace of undoing",
		"'Petrol bomb' is another name for 'Molotov Cocktail'",
		"Due to anti-German sentiment in the United States, sauerkraut was renamed 'victory cabbage' throughout the duration of World War I",
		"Marcel Proust's seven-volume novel In Search of Lost Time was written from 1909-1922 and contains over 2000 characters",
		"A standard daiquiri consists of rum, lime juice and sugar",
		"The fifth A on a piano has a standard tuning of 440 hertz",
		"Chick-fil-A has a company-wide policy of keeping locations closed on Sundays in accordance with Christian ideals embedded in the company's corporate promise statement",
		"The world record for most mayonnaise eaten in eight minutes is eight pounds, heldby Oleg Zhornitskiy",
		"A cacograph is a deliberate misspelling of a word for comical effect",
		"In March of 2010, an official scientific council determined that an asteroid collision was responsible for the mass-extinction at the end of Mesozoic Era",
		"'The Himalayas,' an 18-hole course of putting greens in the town of St. Andrews, Scotland, is considered to be the first miniature golf course",
		"An egg tooth is used by avian and reptilian offspring to break through an egg during hatching",
		"Churchill Downs is the host track of the Kentucky Derby",
		"The character of James Bond was created by Ian Fleming",
		"Angel Falls in Venezuela is the world's highest waterfall",
		"'Shin splints' is the layman's term for medial tibial stress syndrome",
		"In cooking, the use of a salt crust helps to internally steam a food item and give it the outward texture of a roasted dish",
		"The Football War was a four-day war fought by El Salvador and Honduras in 1969",
		"Merv Griffin has earned over $70 million in royalties for the composition 'Think!' featured in the game show Jeopardy!",
		"The original voice of Meowth in the animated children's show Pokemon was Nathan Price",
		"By general scientific consensus, MSG has been shown to not be a 'significant health hazard'-though some individuals experience minor health symptoms after consumption due to the placebo effect",
		"'Bollocks' was ranked as the eighth-most offensive word in the English language according to a 2006 study by BBC",
		"The first Greyhound race in Great Britain took place on July 24, 1926 at Belle Vue Stadium",
		"Blue light has a wavelength of 440 to 490 nanometers",
		"Mushrooms that have been exposed to UV light are the only natural vegetarian source of vitamin D",
		"You will take years off of your own life simply by worrying about things that are out of your control",
		"Your government knows more about you than you know about yourself",
		"Most people die a thousand spiritual deaths before the arbitrary corporeal one happens",
		"You are not the exception",
		"All the world's a stage, and your loved ones are just actors",
		"You are wasting life in the unfounded belief that your time accrues interest",
		"The existence of a 'soul' has never been proven for a reason",
		"You have no influence over your future",
		"They are just not that into you",
		"The problem is not the rest of the world; the problem is you",
		"Your offspring will be just as mundane as you",
		"You will hurt your foot",
		"Everyone you know and everything you cherish is all drifting apart like exiles on driftwood and there is absolutely nothing you can do about it",
		"Society does not include you",
		"Your mind is coming undone the more you think about your mind coming undone",
		"You are trapped in a routine of consumption and excretion that will only end when you finally end it all",
		"Your children will grow up in an unforgiving and indifferent world",
		"Your recent increase in frequency of sleepless nights is simply a sign that you will never again see what you once dreamt of",
		"Every change is for the worse",
		"To die for your country is worse than to die for nothing",
		"Human beings are the disease",
		"Let go of the rail, let go of the rail",
		"'Blog' is a portmanteu of the words 'web' and 'log'",
		"The four types of tennis court surfaces classified for used in professional play are Clay, Hard, Grass and Carpet",
		"Men on average find women with a .7:1 waist-hip ratio the most physically attractive overall",
		"The domestic cat is classified as an invasive species",
		"The yolk of an egg and its germinal disc is a single cell",
		"Lonesome George is the name of the last known remaining Pinta Island Tortoise",
		"Venison is deer meat",
		"A 'Two-Spirit' is a transgender person who identifies as having both a male and a female gender",
		"The 'Parmigiano-Reggiano' name for cheese is a trademarked in Europe",
		"After water and tea, beer is the third most popular drink worldwide",
		"Maple syrup is sap from the maple tree",
		"Polaris is the name of the earth's current northern pole star",
		"The Gulf Stream is part of the ocean current system called the North Atlantic Gyre",
		"The Herring Gull is the most common gull in Asia, North America and western Europe",
		"The Starbucks Corporation is the largest coffeehouse company worldwide",
		"Carnival is a pre-lentin worldwide festival celebrated with street parties and parades",
		"The Dutch oliebol is a deep-fried ball of dough coated with powdered sugar",
		"The card game Uno was created in 1971",
		"The rent for normally landing on a railroad in Monopoly when another player owns all four is $200",
		"TaB soda is a product of the Coca Cola Company",
		"You have done nothing right",
		"You will die alone",
		"Your words have no weight in the world",
		"If you take a stand for your foolish beliefs, no one will stand behind you",
		"Those in charge see you as no more danger to them than a phytoplankton",
		"Your name will be placed on a gravestone and then quickly forgotten",
		"Your friends keep you around because it is amusing when you fail in your pursuits",
		"You are 999,999 in a million",
		"Someday soon the novelty of your personality will wear off and people will see you for the bore that you really are",
		"You make as much impact on the world as a ping pong ball on a brick wall",
		"No one's life would be worse without you in it",
		"You are but a placeholder for the person who comes after you",
		"They only laugh at your jokes out of pity",
		"Every path you choose will lead to equal disappointment",
		"No one is keeping tally of your feeble existence but you",
		"The causes you have taken up are laughable",
		"A garbage bag is just as good a home to a dead man",
		"The worms look at you as a number waiting to be called out in the deli line",
		"Nothing will make up for the shameful things you have done"}
)

// Load returns the Plugin object.
func Load() onelib.Plugin {
	return new(FactCrowPlugin)
}

func factCrow(msg onelib.Message, sender onelib.Sender) {
	randn := rand.Intn(len(factCrowAnswers))
	text := factCrowAnswers[randn] + "."
	sender.Location().SendText(text)
}

// FactCrowPlugin is an object for satisfying the Plugin interface.
type FactCrowPlugin int

// Name returns the name of the plugin, usually the filename.
func (eb *FactCrowPlugin) Name() string {
	return NAME
}

// LongName returns the display name of the plugin.
func (eb *FactCrowPlugin) LongName() string {
	return LONGNAME
}

// Version returns the version of the plugin, usually in the format of "v0.0.0".
func (eb *FactCrowPlugin) Version() string {
	return VERSION
}

// Implements returns a map of commands and monitor the plugin implements.
func (eb *FactCrowPlugin) Implements() (map[string]onelib.Command, *onelib.Monitor) {
	return map[string]onelib.Command{"fc": factCrow, "fact": factCrow}, nil
}

// Remove is necessary to satisfy the Plugin interface, it does nothing.
func (eb *FactCrowPlugin) Remove() {
}
