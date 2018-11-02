package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
)

type Comic struct {
	Num        int    `json:"num"`
	Transcript string `json:"transcript"`
}

const (
	comicURL  = "https://xkcd.com/%d/info.0.json"
	pathDB    = "db.json"
	maxComics = 1000
)

func requestComic(num int, ch chan<- *Comic) {
	u := fmt.Sprintf(comicURL, num)

	log.Println("Getting URL", u)
	resp, err := http.Get(u)
	if err != nil {
		log.Println(err)
		ch <- nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("bad response status code: %v", resp.StatusCode)
		ch <- nil
	}

	var comic Comic
	if err := json.NewDecoder(resp.Body).Decode(&comic); err != nil {
		log.Println(err)
		ch <- nil
	}

	log.Printf("Succesfully got %+v", comic.Num)
	ch <- &comic
}

func saveToDB(comics []*Comic) error {
	if err := os.MkdirAll(path.Dir(pathDB), 0666); err != nil {
		return err
	}

	log.Println("Opening file", pathDB)
	f, err := os.OpenFile(pathDB, os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	log.Println("Writing json to file", pathDB)
	if err := json.NewEncoder(f).Encode(comics); err != nil {
		return err
	}

	log.Println("Succesfully writed to file", pathDB)
	return nil
}

func buildDB() error {
	var comics []*Comic
	ch := make(chan *Comic)
	for i := 1; i <= maxComics; i++ {
		go requestComic(i, ch)
	}
	for i := 1; i <= maxComics; i++ {
		if comic := <-ch; comic != nil {
			comics = append(comics, comic)
		}
	}

	if err := saveToDB(comics); err != nil {
		return err
	}

	return nil
}

func searchDB(num int) (*Comic, error) {
	log.Println("Opening file", pathDB)
	f, err := os.Open(pathDB)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var comics []Comic
	if err := json.NewDecoder(f).Decode(&comics); err != nil {
		return nil, err
	}
	for _, c := range comics {
		if c.Num == num {
			return &c, nil
		}
	}

	return nil, fmt.Errorf("%d not found in database, try to rebuild it", num)
}

func main() {
	buildDatabase := flag.Bool("b", false, "build the database")
	requestedNumber := flag.Int("s", 0, "number of comic to search")
	flag.Parse()

	if *buildDatabase == true {
		if err := buildDB(); err != nil {
			log.Println("Some error while building:", err)
			os.Exit(1)
		}
	} else if *requestedNumber > 0 {
		comic, err := searchDB(*requestedNumber)
		if err != nil {
			log.Println("Error while searching in DB:", err)
			os.Exit(1)
		}
		fmt.Println("Succesfully found:", comic.Transcript)
	} else {
		fmt.Println("Please specify some options, -h for help")
	}

}
