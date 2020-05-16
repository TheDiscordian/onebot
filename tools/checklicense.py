#!/usr/bin/python
# Copyright (c) 2020, The OneBot Contributors. All rights reserved.
import os

# Recursively search every directory downwards, skipping anything beginning with ".", returning all files ending in ".go"
def getallgofiles(path):
    files = []
    for i in os.listdir(path):
        if i.split('/')[-1][0] == '.':
            continue
        if os.path.isfile(os.path.join(path, i)):
            if len(i) > 3 and i[-3:] == '.go':
                files.append(os.path.join(path, i))
        else:
            files.extend(getallgofiles(os.path.join(path, i)))
    return files

copyright = "// " + open('LICENSE').readlines()[0]
gofiles = getallgofiles('.')

for i in gofiles:
    f = open(i).readlines()
    try:
        f.index(copyright)
    except:
        print("No license found on %s, updating..." % i.split('/')[-1])
        out = [copyright+'\n'] # add an extra line so golang doesn't confuse the license for a package comment
        out.extend(f)
        open(i, 'w').write(''.join(out))

