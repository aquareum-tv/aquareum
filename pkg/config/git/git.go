package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/google/uuid"
)

func main() {
	err := makeGit()
	if err != nil {
		panic(err)
	}
}

var tmpl = `
package main

var Version = "%s"
var BuildTime = "%d"
var UUID = "%s"
`

func makeGit() error {
	output := flag.String("o", "", "file to output to")
	version := flag.Bool("v", false, "just print version")
	env := flag.Bool("env", false, "print a bunch of useful environment variables")

	flag.Parse()
	r, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return err
	}

	// ... retrieving the HEAD reference
	ref, err := r.Head()
	if err != nil {
		return err
	}
	c, err := r.CommitObject(ref.Hash())
	if err != nil {
		return err
	}

	ts := c.Author.When.Unix()
	rander := rand.New(rand.NewSource(ts))
	u, err := uuid.NewV7FromReader(rander)
	if err != nil {
		return err
	}
	g, err := PlainOpen(".")
	if err != nil {
		return err
	}
	desc, err := g.Describe(ref)
	if err != nil {
		return err
	}
	var out string
	if *version {
		out = desc
	} else if *env {
		CI_API_V4_URL := os.Getenv("CI_API_V4_URL")
		CI_PROJECT_ID := os.Getenv("CI_PROJECT_ID")
		BRANCH := os.Getenv("BRANCH")
		outMap := map[string]string{}
		outMap["AQUAREUM_BRANCH"] = BRANCH
		outMap["AQUAREUM_VERSION"] = desc
		// https://git.aquareum.tv/api/v4/projects/1/packages/generic/$(BRANCH)/aquareum-v0.0.9-8650d0fa-linux-arm64.tar.gz
		for _, arch := range []string{"amd64", "arm64"} {
			k := fmt.Sprintf("AQUAREUM_URL_%s", strings.ToUpper(arch))
			v := fmt.Sprintf("%s/projects/%s/packages/generic/%s/%s/aquareum-%s-linux-%s.tar.gz", CI_API_V4_URL, CI_PROJECT_ID, BRANCH, desc, desc, arch)
			outMap[k] = v
		}
		for k, v := range outMap {
			out = out + fmt.Sprintf("%s=%s\n", k, v)
		}
	} else {
		out = fmt.Sprintf(tmpl, desc, ts, u)
	}

	if *output != "" {
		os.WriteFile(*output, []byte(out), 0644)
	} else {
		fmt.Print(out)
	}
	// 	if *version {
	// 	} else {
	// 		out := fmt.Sprintf(tmpl, desc, ts, u)
	// 		os.WriteFile(*output, []byte(out), 0644)
	// 	}
	// } else if *version {
	// 	fmt.Print(desc)
	// } else {
	// 	fmt.Printf("%d %s %s", ts, u, desc)
	// }
	return nil
}

// Git struct wrapps Repository class from go-git to add a tag map used to perform queries when describing.
type Git struct {
	TagsMap map[plumbing.Hash]*plumbing.Reference
	*git.Repository
}

// PlainOpen opens a git repository from the given path. It detects if the
// repository is bare or a normal one. If the path doesn't contain a valid
// repository ErrRepositoryNotExists is returned
func PlainOpen(path string) (*Git, error) {
	r, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})
	return &Git{
		make(map[plumbing.Hash]*plumbing.Reference),
		r,
	}, err
}

func (g *Git) getTagMap() error {
	tags, err := g.Tags()
	if err != nil {
		return err
	}

	err = tags.ForEach(func(t *plumbing.Reference) error {
		h, err := g.ResolveRevision(plumbing.Revision(t.Name()))
		if err != nil {
			return err
		}
		g.TagsMap[*h] = t
		return nil
	})

	return err
}

// Describe the reference as 'git describe --tags' will do
func (g *Git) Describe(reference *plumbing.Reference) (string, error) {

	// Fetch the reference log
	cIter, err := g.Log(&git.LogOptions{
		// From:  reference.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return "", err
	}

	// Build the tag map
	err = g.getTagMap()
	if err != nil {
		return "", err
	}

	// Search the tag
	var tag *plumbing.Reference
	var count int
	err = cIter.ForEach(func(c *object.Commit) error {
		t, ok := g.TagsMap[c.Hash]
		if ok {
			tag = t
			return storer.ErrStop
		}
		count++
		return nil
	})
	if err != nil {
		return "", err
	}
	head, err := g.Head()
	if err != nil {
		return "", err
	}
	if count == 0 {
		return fmt.Sprint(tag.Name().Short()), nil
	} else {
		return fmt.Sprintf("%s-%s",
			tag.Name().Short(),
			head.Hash().String()[0:8],
		), nil
	}
}
