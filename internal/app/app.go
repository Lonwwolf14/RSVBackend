package app

import (
	"html/template"
	"rsvbackend/internal/database"
	"sync"

	"github.com/gorilla/sessions"
)

// AppState holds the application-wide state and dependencies
type AppState struct {
	DB        database.QueriesInterface // Database queries
	Store     *sessions.CookieStore     // Session store for authentication
	Templates *template.Template        // Loaded templates
	Node      *Node                     // Distributed system node state
}

// Node represents the state of a node in the distributed system
type Node struct {
	ID       string
	Clock    int64
	Peers    []string
	InCS     bool // This node is in critical section
	AnyCS    bool // Any node is in critical section
	Requests []Request
	Mutex    sync.Mutex
}

type Request struct {
	NodeID    string
	Timestamp int64
	UserID    string
}

// NewAppState initializes a new AppState with the given dependencies
func NewAppState(
	db database.QueriesInterface,
	store *sessions.CookieStore,
	templates *template.Template,
	nodeID string,
	peers []string,
) *AppState {
	return &AppState{
		DB:        db,
		Store:     store,
		Templates: templates,
		Node: &Node{
			ID:    nodeID,
			Clock: 0,
			Peers: peers,
		},
	}
}
