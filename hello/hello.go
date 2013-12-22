package hello

import (
    "html/template"
    "net/http"
    "time"

    "appengine"
    "appengine/datastore"
)

type Place struct {
    Location string
    Note  string
    Date  time.Time
}

func init() {
    http.HandleFunc("/", root)
    http.HandleFunc("/place", place)
}

// placeKey returns the key used for all guestbook entries.
func placeKey(c appengine.Context) *datastore.Key {
    // The string "default_place" here could be varied to have multiple places.
    return datastore.NewKey(c, "Place", "mara_place", 0, nil)
}

func root(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    // Ancestor queries, as shown here, are strongly consistent with the High
    // Replication Datastore. Queries that span entity groups are eventually
    // consistent. If we omitted the .Ancestor from this query there would be
    // a slight chance that Greeting that had just been written would not
    // show up in a query.
    q := datastore.NewQuery("Place").Ancestor(placeKey(c)).Order("-Date").Limit(10)
    places := make([]Place, 0, 10)
    if _, err := q.GetAll(c, &places); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if err := placesTemplate.Execute(w, places); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

var placesTemplate = template.Must(template.New("places").Parse(placeTemplateHtml))

const placeTemplateHtml = `
<html>
  <body>
    {{range .}}
      {{with .Location}}
        <p><b>{{.}}</b> and </p>
      {{end}}
      <pre>{{.Note}}</pre>
    {{end}}
    <form action="/place" method="post">
      <div><textarea name="location" rows="3" cols="60"></textarea></div>
      <div><textarea name="note" rows="3" cols="60"></textarea></div>
      <div><input type="submit" value="Submit Place"></div>
    </form>
  </body>
</html>
`

func place(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    g := Place{
        Location: r.FormValue("location"),
        Note: r.FormValue("note"),
        Date:    time.Now(),
    }
    // We set the same parent key on every Place entity to ensure each Place
    // is in the same entity group. Queries across the single entity group
    // will be consistent. However, the write rate to a single entity group
    // should be limited to ~1/second.
    key := datastore.NewIncompleteKey(c, "Place", placeKey(c))
    _, err := datastore.Put(c, key, &g)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, "/", http.StatusFound)
}
