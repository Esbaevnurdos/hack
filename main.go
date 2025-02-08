package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/gorilla/handlers"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/Esbaevnurdos/hack/docs" // Change 'your_project' to match your module name

	"github.com/gorilla/mux"
)

// Place represents a tourist place
type Place struct {
	ID          int      `json:"id"`
	PlaceName   string   `json:"placeName"`
	Rating      float64  `json:"rating"`
	Description string   `json:"description"`
	PhotoURLs   []string `json:"photoURLs"`
	Comments    []string `json:"comments"`
	Latitude    float64  `json:"latitude"`
	Longitude   float64  `json:"longitude"`
}


var (
	places   []Place
	mutex    sync.Mutex
	jsonFile = "places.json"
)

// @title Taraz Places API
// @version 1.0
// @description API for managing tourist places in Taraz
// @host hack-zjzn.onrender.com
// @BasePath /

func main() {
	loadPlaces()

	r := mux.NewRouter()

	r.HandleFunc("/places", getPlaces).Methods("GET")
	r.HandleFunc("/places/{id}", getPlaceByID).Methods("GET")
	r.HandleFunc("/places", createPlace).Methods("POST")
	r.HandleFunc("/places/{id}", updatePlace).Methods("PUT")
	r.HandleFunc("/places/{id}", deletePlace).Methods("DELETE")

r.HandleFunc("/places/{id}/rate", ratePlace).Methods("POST")
r.HandleFunc("/places/{id}/comment", addComment).Methods("POST")
r.HandleFunc("/places/{id}/photo", addPhoto).Methods("POST")

	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	r.Use(handlers.CORS(
    handlers.AllowedOrigins([]string{"*"}), // Adjust allowed origins as needed
    handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
    handlers.AllowedHeaders([]string{"Content-Type"}),
))

	fmt.Println("Server running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func loadPlaces() {
    file, err := os.ReadFile(jsonFile)
    if err != nil {
        log.Println("Warning: Could not load places.json. Starting with an empty list.")
        places = []Place{} // Ensure places is initialized
        return
    }

    if err := json.Unmarshal(file, &places); err != nil {
        log.Println("Error parsing places.json:", err)
        places = []Place{} // Reset in case of failure
        return
    }
}


func savePlaces() {
	data, _ := json.MarshalIndent(places, "", "  ")
	os.WriteFile(jsonFile, data, 0644)
}

// @Summary Get all places
// @Description Returns a list of all places
// @Produce json
// @Success 200 {array} Place
// @Router /places [get]
func getPlaces(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(places)
}

// @Summary Get place by ID
// @Description Returns a place by its ID
// @Param id path int true "Place ID"
// @Produce json
// @Success 200 {object} Place
// @Failure 404 {string} string "Place not found"
// @Router /places/{id} [get]
func getPlaceByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
if err != nil {
    http.Error(w, "Invalid ID format", http.StatusBadRequest)
    return
}


	for _, p := range places {
		if p.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(p)
			return
		}
	}

	http.Error(w, "Place not found", http.StatusNotFound)
}

// @Summary Create a new place
// @Description Adds a new place to the list
// @Accept json
// @Produce json
// @Param place body Place true "Place data"
// @Success 201 {string} string "Place added"
// @Router /places [post]
func createPlace(w http.ResponseWriter, r *http.Request) {
	var newPlace Place
	if err := json.NewDecoder(r.Body).Decode(&newPlace); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate Latitude and Longitude
	if newPlace.Latitude < -90 || newPlace.Latitude > 90 {
		http.Error(w, "Invalid latitude value", http.StatusBadRequest)
		return
	}
	if newPlace.Longitude < -180 || newPlace.Longitude > 180 {
		http.Error(w, "Invalid longitude value", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	newPlace.ID = len(places) + 1
	places = append(places, newPlace)
	savePlaces()
	mutex.Unlock()

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Place added"))
}


// @Summary Update a place
// @Description Updates an existing place by ID
// @Accept json
// @Produce json
// @Param id path int true "Place ID"
// @Param place body Place true "Updated place data"
// @Success 200 {string} string "Place updated"
// @Router /places/{id} [put]
func updatePlace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var updatedData Place
	if err := json.NewDecoder(r.Body).Decode(&updatedData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate Latitude and Longitude
	if updatedData.Latitude < -90 || updatedData.Latitude > 90 {
		http.Error(w, "Invalid latitude value", http.StatusBadRequest)
		return
	}
	if updatedData.Longitude < -180 || updatedData.Longitude > 180 {
		http.Error(w, "Invalid longitude value", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	for i, p := range places {
		if p.ID == id {
			// Update fields
			places[i].PlaceName = updatedData.PlaceName
			places[i].Rating = updatedData.Rating
			places[i].Description = updatedData.Description
			places[i].PhotoURLs = updatedData.PhotoURLs
			places[i].Comments = updatedData.Comments
			places[i].Latitude = updatedData.Latitude
			places[i].Longitude = updatedData.Longitude

			savePlaces()
			w.Write([]byte("Place updated"))
			return
		}
	}

	http.Error(w, "Place not found", http.StatusNotFound)
}



// @Summary Delete a place
// @Description Deletes a place by ID
// @Param id path int true "Place ID"
// @Success 200 {string} string "Place deleted"
// @Router /places/{id} [delete]
func deletePlace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	mutex.Lock()
	for i, p := range places {
		if p.ID == id {
			places = append(places[:i], places[i+1:]...)
			savePlaces()
			mutex.Unlock()
			w.Write([]byte("Place deleted"))
			return
		}
	}
	mutex.Unlock()

	http.Error(w, "Place not found", http.StatusNotFound)
}

// @Summary Rate a place
// @Description Updates the rating of a place
// @Param id path int true "Place ID"
// @Param rating query float64 true "New rating"
// @Success 200 {string} string "Rating updated"
// @Router /places/{id}/rate [post]
func ratePlace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid place ID", http.StatusBadRequest)
		return
	}

	// Read rating from JSON body
	var req struct {
		Rating float64 `json:"rating"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	for i, p := range places {
		if p.ID == id {
			places[i].Rating = req.Rating
			savePlaces()
			w.Write([]byte("Rating updated"))
			return
		}
	}

	http.Error(w, "Place not found", http.StatusNotFound)
}



// @Summary Add comment
// @Description Adds a comment to a place
// @Param id path int true "Place ID"
// @Param comment body string true "Comment text"
// @Success 200 {string} string "Comment added"
// @Router /places/{id}/comment [post]
func addComment(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    placeID, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    var comment struct {
        Text string `json:"text"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    mutex.Lock()
    defer mutex.Unlock()

    for i, p := range places {
        if p.ID == placeID {
            places[i].Comments = append(places[i].Comments, comment.Text)
            savePlaces()
            w.Write([]byte("Comment added"))
            return
        }
    }

    http.Error(w, "Place not found", http.StatusNotFound)
}


// @Summary Add photo
// @Description Adds a photo URL to a place
// @Param id path int true "Place ID"
// @Param photo body struct{URL string `json:"url"`} true "Photo URL"
// @Success 200 {string} string "Photo added"
// @Router /places/{id}/photo [post]
func addPhoto(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid place ID", http.StatusBadRequest)
		return
	}

	var photo struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&photo); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock() // Always unlock

	for i, p := range places {
		if p.ID == id {
			places[i].PhotoURLs = append(places[i].PhotoURLs, photo.URL)
			savePlaces()
			w.Write([]byte("Photo added"))
			return
		}
	}

	http.Error(w, "Place not found", http.StatusNotFound)
}


