--- EDIT ---

Changed this project to Go instead of Python. In the end this will have four services.

1 - postgresql database with two tables; one for current movies, one for recommended movies
2 - web service that is the frontend of this postgres data
    * HTTP Post requests to /movies/ will add data for recommended movie
    * HTTP Get request to /movies/ will give a list of all recommended movie titles and IMDb ID'
    * HTTP Get request to /movie/# where # is a specific ID tag will display information about the recommended movie, including other movies that recommended it
3 - IMDb scraping API. Scrapes IMDb page for movie facts and stores them in a struct such as: title, rating, genre, cast(first 5), recommendations, director
4 - Plex API call to gather library list of movies and provide IMDb tag and other information

ultimately service4 should be called to gather a list of movies, if there are new movies pass those IMDb tags to service3.
Service3 gathers the list of movies IMDb recommends and then recursively passes those tags back to itself (service3) to gather movie facts.
Service3 then can pass this new recommended movie's information via Post request to service2 in order to populate the tables for service1

there will most likely be another service that does calculations and manipulation of weighted values for recommendation on the movies

--- EDIT ---

layout idea

file 1 - grab IMDb movieID's from plex movie files, put them in dynamoDB table. grab movieID, title, cast(5), director, gengre

file 2 - look at dynamo table for movies we have, fetch IMDb's recommended. save those and grab their title, cast(5), director, gengre, movieID recommended from. save to another dynamo table?

file 3 - open both tables and run a count of each of the variables. do matrix factorization based on previous variables using counts of each for the weight, more times they show up in our list higher their weight more recommended the movie is. double recommended_from variable so movies recommended by multiple of our movies are higher on the list over everything else