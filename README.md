layout idea

file 1 - grab IMDb movieID's from plex movie files, put them in dynamoDB table. grab movieID, title, cast(5), director, gengre

file 2 - look at dynamo table for movies we have, fetch IMDb's recommended. save those and grab their title, cast(5), director, gengre, movieID recommended from. save to another dynamo table?

file 3 - open both tables and run a count of each of the variables. do matrix factorization based on previous variables using counts of each for the weight, more times they show up in our list higher their weight more recommended the movie is. double recommended_from variable so movies recommended by multiple of our movies are higher on the list over everything else