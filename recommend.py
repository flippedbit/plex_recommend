#!/usr/bin/python

from imdb import IMDb
from sys import argv
from sys import exit

if len(argv) > 1:
    movieID = argv[1].strip("t")
else:
    exit(1)

ia = IMDb()
recommend = {}
recs = ia.get_movie_recommendations(movieID)
movie = ia.get_movie(movieID)
cast = movie.get("cast")[0:5]

recommend[movie] = {}
recommend[movie]["cast"] = cast
print recommend
#print "Cast: "
#print cast
#print "Recommendations:"
#for rec in recs["data"]["recommendations"]:
    #print("%s - %s" % (rec, rec.getID()))
