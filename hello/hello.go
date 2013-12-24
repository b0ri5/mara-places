package hello

import (
    "net/http"
    "time"
    "encoding/json"

    "appengine"
    "appengine/datastore"
)

type Place struct {
    Location string
    Note  string
    Date  time.Time
}

func init() {
    http.HandleFunc("/worldlove", worldlove)
}

// placeKey returns the key used for all guestbook entries.
func placeKey(c appengine.Context) *datastore.Key {
    // The string "default_place" here could be varied to have multiple places.
    return datastore.NewKey(c, "Place", "mara_place", 0, nil)
}

func worldlove(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
	getPlace(w, r)
    } else if r.Method == "POST" {
	putPlace(w, r)
    }
}

func getPlace(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    // Ancestor queries, as shown here, are strongly consistent with the High
    // Replication Datastore. Queries that span entity groups are eventually
    // consistent. If we omitted the .Ancestor from this query there would be
    // a slight chance that Greeting that had just been written would not
    // show up in a query.
    limit := 1000
    q := datastore.NewQuery("Place").Ancestor(placeKey(c)).Order("Date").Limit(limit)
    places := make([]Place, 0, limit)
    if _, err := q.GetAll(c, &places); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    b, err := json.Marshal(places)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    if _, err := w.Write(b); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
}

func putPlace(w http.ResponseWriter, r *http.Request) {
    place := Place{}
    if err := json.NewDecoder(r.Body).Decode(&place); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    c := appengine.NewContext(r)
    // We set the same parent key on every Place entity to ensure each Place
    // is in the same entity group. Queries across the single entity group
    // will be consistent. However, the write rate to a single entity group
    // should be limited to ~1/second.
    key := datastore.NewIncompleteKey(c, "Place", placeKey(c))
    _, err := datastore.Put(c, key, &place)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
}
