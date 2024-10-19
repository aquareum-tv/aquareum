package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"sort"
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

var tmplJS = `
export const version = "%s";
export const buildTime = "%d";
export const uuid = "%s";
`

func gitlabURL() string {
	CI_API_V4_URL := os.Getenv("CI_API_V4_URL")
	CI_PROJECT_ID := os.Getenv("CI_PROJECT_ID")
	return fmt.Sprintf("%s/projects/%s", CI_API_V4_URL, CI_PROJECT_ID)
}

func gitlab(suffix string, dest any) {
	u := fmt.Sprintf("%s%s", gitlabURL(), suffix)

	req, err := http.Get(u)
	if err != nil {
		panic(err)
	}
	if err := json.NewDecoder(req.Body).Decode(dest); err != nil {
		panic(err)
	}
}

func gitlabList(suffix string) []map[string]any {
	var result []map[string]any
	gitlab(suffix, &result)
	return result
}

func gitlabDict(suffix string) map[string]any {
	var result map[string]any
	gitlab(suffix, &result)
	return result
}

func makeGit() error {
	output := flag.String("o", "", "file to output to")
	version := flag.Bool("v", false, "just print version")
	env := flag.Bool("env", false, "print a bunch of useful environment variables")
	doBranch := flag.Bool("branch", false, "print branch")
	doRelease := flag.Bool("release", false, "print release json file")
	javascript := flag.Bool("js", false, "print code in javascript format")

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
	} else if *doBranch {
		out = branch()
	} else if *env {
		AQUAREUM_BRANCH := branch()
		outMap := map[string]string{}
		outMap["AQUAREUM_BRANCH"] = AQUAREUM_BRANCH
		outMap["AQUAREUM_VERSION"] = desc
		outMap["AQUAREUM_BRANCH"] = AQUAREUM_BRANCH
		for _, arch := range []string{"amd64", "arm64"} {
			k := fmt.Sprintf("AQUAREUM_URL_LINUX_%s", strings.ToUpper(arch))
			v := fmt.Sprintf("%s/packages/generic/%s/%s/aquareum-%s-linux-%s.tar.gz", gitlabURL(), AQUAREUM_BRANCH, desc, desc, arch)
			outMap[k] = v
			macK := fmt.Sprintf("AQUAREUM_URL_DARWIN_%s", strings.ToUpper(arch))
			macV := fmt.Sprintf("%s/packages/generic/%s/%s/aquareum-%s-darwin-%s.zip", gitlabURL(), AQUAREUM_BRANCH, desc, desc, arch)
			outMap[macK] = macV
		}
		outMap["AQUAREUM_URL_WINDOWS_AMD64"] = fmt.Sprintf("%s/packages/generic/%s/%s/aquareum-%s-windows-amd64.exe", gitlabURL(), AQUAREUM_BRANCH, desc, desc)
		for k, v := range outMap {
			out = out + fmt.Sprintf("%s=%s\n", k, v)
		}
	} else if *doRelease {
		outMap := map[string]any{}
		outMap["name"] = desc
		outMap["tag-name"] = desc
		pkgs := gitlabList(fmt.Sprintf("/packages?order_by=created_at&sort=desc&package_name=%s", branch()))
		id := pkgs[0]["id"].(float64)
		pkgFiles := gitlabList(fmt.Sprintf("/packages/%d/package_files", int(id)))
		outFiles := []string{}
		sort.Slice(pkgFiles, func(i, j int) bool {
			s1 := pkgFiles[i]["file_name"].(string)
			s2 := pkgFiles[j]["file_name"].(string)
			return s1 < s2
		})
		for _, file := range pkgFiles {
			fileJson := map[string]string{
				"name": file["file_name"].(string),
				"url":  fmt.Sprintf("%s/packages/generic/%s/%s/%s", gitlabURL(), branch(), desc, file["file_name"].(string)),
			}
			bs, err := json.Marshal(fileJson)
			if err != nil {
				return err
			}
			outFiles = append(outFiles, string(bs))
		}
		outMap["assets-link"] = outFiles
		changelog := gitlabDict(fmt.Sprintf("/repository/changelog?version=%s", desc))
		outMap["description"] = changelog["notes"]
		bs, err := json.MarshalIndent(outMap, "", "  ")
		if err != nil {
			return err
		}
		out = string(bs)
	} else if *javascript {
		out = fmt.Sprintf(tmplJS, desc, ts, u)
	} else {
		out = fmt.Sprintf(tmpl, desc, ts, u)
	}

	if *output != "" {
		os.WriteFile(*output, []byte(out), 0644)
	} else {
		fmt.Print(out)
	}
	return nil
}

func branch() string {
	CI_COMMIT_TAG := os.Getenv("CI_COMMIT_TAG")
	CI_COMMIT_BRANCH := os.Getenv("CI_COMMIT_BRANCH")
	if CI_COMMIT_TAG != "" {
		return "latest"
	} else if CI_COMMIT_BRANCH != "" {
		return strings.Replace(CI_COMMIT_BRANCH, "/", "-", -1)
	} else {
		panic("CI_COMMIT_TAG and CI_COMMIT_BRANCH undefined, can't get branch")
	}
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
