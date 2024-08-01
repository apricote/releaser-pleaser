package rp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

const (
	CommitSearchDepth = 50 // TODO: Increase
	GitRemoteName     = "origin"
)

type Commit struct {
	Hash    string
	Message string
}

type Tag struct {
	Hash string
	Name string
}

func CloneRepo(ctx context.Context, cloneURL, branch string, auth transport.AuthMethod) (*git.Repository, error) {
	dir, err := os.MkdirTemp("", "releaser-pleaser.*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory for repo clone: %w", err)
	}

	// TODO: Log tmpdir
	fmt.Printf("Clone tmpdir: %s\n", dir)
	repo, err := git.PlainCloneContext(ctx, dir, false, &git.CloneOptions{
		URL:           cloneURL,
		RemoteName:    GitRemoteName,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		SingleBranch:  false,
		Auth:          auth,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	return repo, nil
}

func ReleasableCommits(repo *git.Repository) ([]Commit, *Tag, error) {

	ref, err := repo.Head()
	if err != nil {
		return nil, nil, err
	}

	iter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, nil, err
	}

	tags, err := buildTagRefMap(repo)
	if err != nil {
		return nil, nil, err
	}

	commits := make([]Commit, 0)
	var tag *Tag
	for {
		commit, err := iter.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, nil, err
		}

		if tagRef, exists := tags[commit.Hash]; exists {
			// We found the nearest tag, return results
			tagName, _ := strings.CutPrefix(tagRef.Name().String(), "refs/tags/")

			tag = &Tag{
				Hash: tagRef.Hash().String(),
				Name: tagName,
			}
			break
		}

		commits = append(commits, Commit{
			Hash:    commit.Hash.String(),
			Message: commit.Message,
		})
	}

	// We discover the commits from HEAD, but want to process them in "chronological" order
	slices.Reverse(commits)

	return commits, tag, nil
}

// From go-git PR
func buildTagRefMap(r *git.Repository) (map[plumbing.Hash]*plumbing.Reference, error) {
	iter, err := r.Tags()
	if err != nil {
		return nil, err
	}
	tags := map[plumbing.Hash]*plumbing.Reference{}
	for {
		t, err := iter.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		to, err := r.TagObject(t.Hash())
		if errors.Is(err, plumbing.ErrObjectNotFound) {
			// t is a lightweight tag
			tags[t.Hash()] = t
		} else if err != nil {
			return nil, err
		} else {
			// t is an annotated tag
			tags[to.Target] = t
		}
	}
	return tags, nil
}
