package main

import (
	"flag"
	"fmt"
	"github.com/peterh/liner"
	"github.com/veonik/squircy2/data"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	oldsquircy2 "./internal/oldsquircy2"
	olddata "./internal/oldsquircy2/data"
	"github.com/veonik/squircy2"

	"gopkg.in/mattes/go-expand-tilde.v1"
)

var oldRootPathFlag = flag.String("old-root-path", "~/.squircy2-old", "Specify a custom root path.")
var newRootPathFlag = flag.String("new-root-path", "~/.squircy2", "Specify a custom root path.")

var Version = "dev"
var GoVersion = runtime.Version()

func printVersion() {
	fmt.Printf("squIRcy2 version %s (%s)", Version, GoVersion)
}

type hlpr struct{
	Name string
	Records []map[string]interface{}
	Errors []error
	Inserted []int
}

func (h *hlpr) insert(database *data.DB) {
	repo := database.Use(h.Name)
	for _, v := range h.Records {
		id, err := repo.Insert(v)
		if err != nil {
			h.Errors = append(h.Errors, err)
		} else {
			h.Inserted = append(h.Inserted, id)
		}
	}
}

func main() {
	flag.Usage = func() {
		printVersion()
		fmt.Println()
		fmt.Println("Usage: squircy-dump [-oldRoot-path <config oldRoot>]")
		fmt.Println()
		flag.PrintDefaults()
	}
	flag.Parse()

	oldRoot, err := tilde.Expand(*oldRootPathFlag)
	if err != nil {
		panic(err)
	}

	newRoot, err := tilde.Expand(*newRootPathFlag)
	if err != nil {
		panic(err)
	}

	oldies := oldsquircy2.NewManager(oldRoot)

	mgr := squircy2.NewManager(newRoot)

	fmt.Println("exporting data from", oldRoot)

	docs := map[string][]map[string]interface{}{}

	db := oldies.DB()
	for _, colName := range db.AllCols() {
		repo := olddata.NewGenericRepository(db, colName)
		vals := repo.FetchAll()

		for _, v := range vals {
			delete(v, "ID")
			docs[colName] = append(docs[colName], v)
		}

		fmt.Println("found", len(docs[colName]), "documents in collection", colName)
	}

	fmt.Println("export complete, preparing to import into new version")
	fmt.Println("importing data into", newRoot)
	in := liner.NewLiner()
	res, err := in.PromptWithSuggestion("continue? [Yes/no] ", "yes", 0)
	if err != nil {
		fmt.Println(err)
		res = ""
	}
	if len(res) < 1 || strings.ToLower(res)[0] != 'y' {
		fmt.Println("aborting!")
		os.Exit(10)
	}

	for colName, docs := range docs {
		err := os.RemoveAll(filepath.Join(newRoot, "data", colName))
		if err != nil {
			fmt.Println("unable to remove existing folder for collection", colName, err)
			continue
		}

		err = os.MkdirAll(filepath.Join(newRoot, "data", colName), 0755)
		if err != nil {
			fmt.Println("unable to recreate folder for collection", colName, err)
			continue
		}

		h := &hlpr{
			Name: colName,
			Records: docs,
		}

		if _, err := mgr.Invoke(h.insert); err != nil {
			fmt.Println("error converting collection", colName, err)
		}
		if len(h.Errors) > 0 {
			fmt.Println("errors encountered during conversion of collection", colName)
			for _, err := range h.Errors {
				fmt.Println(err)
			}
		}
	}

}
