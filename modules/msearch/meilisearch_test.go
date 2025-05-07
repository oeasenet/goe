package msearch

import (
	"github.com/goccy/go-json"
	"github.com/meilisearch/meilisearch-go"
	"github.com/stretchr/testify/assert"
	"testing"
)

type movie struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Overview    string   `json:"overview"`
	Genres      []string `json:"genres"`
	Poster      string   `json:"poster"`
	ReleaseDate int      `json:"release_date"`
}

var moviesJson string = `[
{"id":2,"title":"Ariel","overview":"Taisto Kasurinen is a Finnish coal miner whose father has just committed suicide and who is framed for a crime he did not commit. In jail, he starts to dream about leaving the country and starting a new life. He escapes from prison but things don't go as planned...","genres":["Drama","Crime","Comedy"],"poster":"https://image.tmdb.org/t/p/w500/ojDg0PGvs6R9xYFodRct2kdI6wC.jpg","release_date":593395200},
{"id":5,"title":"Four Rooms","overview":"It's Ted the Bellhop's first night on the job...and the hotel's very unusual guests are about to place him in some outrageous predicaments. It seems that this evening's room service is serving up one unbelievable happening after another.","genres":["Crime","Comedy"],"poster":"https://image.tmdb.org/t/p/w500/75aHn1NOYXh4M7L5shoeQ6NGykP.jpg","release_date":818467200},
{"id":6,"title":"Judgment Night","overview":"While racing to a boxing match, Frank, Mike, John and Rey get more than they bargained for. A wrong turn lands them directly in the path of Fallon, a vicious, wise-cracking drug lord. After accidentally witnessing Fallon murder a disloyal henchman, the four become his unwilling prey in a savage game of cat & mouse as they are mercilessly stalked through the urban jungle in this taut suspense drama","genres":["Action","Thriller","Crime"],"poster":"https://image.tmdb.org/t/p/w500/rYFAvSPlQUCebayLcxyK79yvtvV.jpg","release_date":750643200},
{"id":11,"title":"Star Wars","overview":"Princess Leia is captured and held hostage by the evil Imperial forces in their effort to take over the galactic Empire. Venturesome Luke Skywalker and dashing captain Han Solo team together with the loveable robot duo R2-D2 and C-3PO to rescue the beautiful princess and restore peace and justice in the Empire.","genres":["Adventure","Action","Science Fiction"],"poster":"https://image.tmdb.org/t/p/w500/6FfCtAuVAW8XJjZ7eWeLibRLWTw.jpg","release_date":233366400},
{"id":12,"title":"Finding Nemo","overview":"Nemo, an adventurous young clownfish, is unexpectedly taken from his Great Barrier Reef home to a dentist's office aquarium. It's up to his worrisome father Marlin and a friendly but forgetful fish Dory to bring Nemo home -- meeting vegetarian sharks, surfer dude turtles, hypnotic jellyfish, hungry seagulls, and more along the way.","genres":["Animation","Family"],"poster":"https://image.tmdb.org/t/p/w500/eHuGQ10FUzK1mdOY69wF5pGgEf5.jpg","release_date":1054252800}
]`

func getMovies() []movie {
	var movies []movie
	_ = json.Unmarshal([]byte(moviesJson), &movies)
	return movies
}

func deleteMovies(search *MSearch) {
	_, _ = search.client.Index("movies").DeleteAllDocuments()
}

func TestAddDocAndGetDoc(t *testing.T) {
	moviesDocs := getMovies()
	search := NewMSearch("http://localhost:7700", "")
	err := search.AddDoc("movies", &moviesDocs[0])
	assert.NoError(t, err)
	response := &movie{}
	hasResult, err := search.GetDoc("movies", "2", response)
	assert.True(t, hasResult)
	assert.NoError(t, err)
	assert.Equal(t, response, &moviesDocs[0])

	defer deleteMovies(search)
}

func TestUpdateDoc(t *testing.T) {
	moviesDocs := getMovies()
	m := &moviesDocs[0]
	search := NewMSearch("http://localhost:7700", "")
	err := search.AddDoc("movies", m)
	assert.NoError(t, err)
	m.ID = 112312
	m.Title = "updated title"

	err = search.UpdateDoc("movies", m)
	assert.NoError(t, err)

	response := &movie{}
	hasResult, err := search.GetDoc("movies", "112312", response)
	assert.True(t, hasResult)
	assert.NoError(t, err)
	assert.Equal(t, response, m)

	defer deleteMovies(search)
}

func TestDeleteDoc(t *testing.T) {
	moviesDocs := getMovies()
	m := &moviesDocs[0]
	search := NewMSearch("http://localhost:7700", "")
	err := search.AddDoc("movies", m)
	assert.NoError(t, err)

	err = search.DelDoc("movies", "2")
	assert.NoError(t, err)

	mov := &movie{}
	hasResult, err := search.GetDoc("movies", "2", mov)
	assert.NoError(t, err)
	assert.Equal(t, mov, &movie{})
	assert.False(t, hasResult)

	defer deleteMovies(search)
}

func TestSearch(t *testing.T) {
	moviesDocs := getMovies()

	search := NewMSearch("http://localhost:7700", "")
	for _, doc := range moviesDocs {
		err := search.AddDoc("movies", &doc)
		assert.NoError(t, err)
	}
	request := meilisearch.SearchRequest{
		HitsPerPage: 5,
		Page:        1,
	}
	response := search.Search("movies", "", &request)

	for i, hit := range response.Hits {
		h, _ := json.Marshal(hit)
		var m movie
		_ = json.Unmarshal(h, &m)

		assert.Equal(t, m, moviesDocs[i])
	}
	defer deleteMovies(search)
}
