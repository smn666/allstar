package webhandler

import (
	"context"
	"fmt"
	"html"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v42/github"
	"github.com/rs/zerolog/log"
)

const secretKey = "FooBar"
const appID = 169668

type WebookHandler struct {
	keyFileName string
}

func HandleWebhooks(privateKeyName string) error {
	w := WebookHandler{keyFileName: privateKeyName}
	http.HandleFunc("/", w.Handle)

	log.Info().Str("key", privateKeyName).Msg("Starting handling with key")

	return http.ListenAndServe(":8080", nil)
}

func (h WebookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))

	payload, err := github.ValidatePayload(r, []byte(secretKey))
	if err != nil {
		log.Error().Err(err).Msg("Got an invalid payload")
		return
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		log.Error().Err(err).Msg("Invalid webhook payload")
		return
	}

	var pr *github.PullRequestReviewEvent

	switch event := event.(type) {
	case *github.PullRequestReviewEvent:
		pr = event
	default:
		log.Warn().Interface("Event", event).Msg("Unknown event")
		return
	}

	log.Info().Interface("PR", pr).Msg("Got a PR event")

	tr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, appID, pr.GetInstallation().GetID(), h.keyFileName)
	if err != nil {
		log.Error().Err(err).Msg("Could not read key")
		return
	}

	client := github.NewClient(&http.Client{Transport: tr})

	// client
	// - determine if user is a member of org OR
	// - has write perms to repo OR
	// - has write perms in org OR

	opt := &github.ListCollaboratorsOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
		Affiliation: "all",
	}

	ctx := context.Background()

	owner := pr.GetRepo().GetOwner().GetLogin()
	repo := pr.GetRepo().GetName()

	users, _, err := client.Repositories.ListCollaborators(ctx, owner, repo, opt)
	if err != nil {
		log.Error().Err(err).Msg("Could not list collaborators")
		return
	}

	var pushers = make(map[string]bool)
	for _, u := range users {
		// NOTE: other permissions: https://docs.github.com/en/rest/reference/collaborators#list-repository-collaborators
		if u.GetPermissions()["push"] {
			log.Debug().Str("pusher", u.GetLogin()).Msg("Found a pusher")
			pushers[u.GetLogin()] = true
		}
	}

	optListReviews := &github.ListOptions{
		PerPage: 100,
	}

	// List of approvers
	reviews, _, err := client.PullRequests.ListReviews(ctx, owner, repo, pr.GetPullRequest().GetNumber(), optListReviews)
	if err != nil {
		log.Error().Err(err).Msg("Could not list reviews")
		return
	}

	// get user names
	points := 0

	if pushers[pr.GetPullRequest().GetUser().GetLogin()] {
		log.Debug().Str("sender", pr.GetPullRequest().GetUser().GetLogin()).Msg("Sender is authorized")
		points++
	}

	// check reviews
	for _, r := range reviews {
		isApprover := pushers[r.GetUser().GetLogin()]

		log.Debug().Str("Review", r.GetUser().GetLogin()).Str("State", r.GetState()).Msg("Found Review")

		if r.GetState() == "APPROVED" && isApprover {
			log.Debug().Str("approver", r.GetUser().GetLogin()).Msg("Found an authorized approver")

			points++
		}
	}

	totalPointsRequired := 2

	log.Info().Int("points", points).Msg("Total points")
	if points >= totalPointsRequired {
		client.Checks.
			log.Info().Msg("Should pass")
	} else {
		log.Info().Msg("Should fail")
	}
}
