// Copyright 2016 Marapongo, Inc. All rights reserved.

package resource

import (
	"github.com/marapongo/mu/pkg/compiler/core"
	"github.com/marapongo/mu/pkg/eval/heapstate"
	"github.com/marapongo/mu/pkg/eval/rt"
	"github.com/marapongo/mu/pkg/graph"
	"github.com/marapongo/mu/pkg/tokens"
	"github.com/marapongo/mu/pkg/util/contract"
)

// Snapshot is a view of a collection of resources in an environment at a point in time.  It describes resources; their
// IDs, names, and properties; their dependencies; and more.  A snapshot is a diffable entity and can be used to create
// or apply an infrastructure deployment plan in order to make reality match the snapshot state.
type Snapshot interface {
	Ctx() *Context                              // fetches the context for this snapshot.
	Pkg() tokens.PackageName                    // the package from which this snapshot came.
	Args() core.Args                            // the arguments used to compile this package.
	Resources() []Resource                      // a topologically sorted list of resources (based on dependencies).
	ResourceByID(id ID, t tokens.Type) Resource // looks up a resource by ID and type.
	ResourceByMoniker(m Moniker) Resource       // looks up a resource by its moniker.
	ResourceByObject(obj *rt.Object) Resource   // looks up a resource by its object.
}

// NewSnapshot creates a snapshot from the given arguments.  Note that resources must be in topologically-sorted
// dependency order, otherwise undefined behavior will result from using the resulting snapshot object.
func NewSnapshot(ctx *Context, pkg tokens.PackageName, args core.Args, resources []Resource) Snapshot {
	return &snapshot{ctx, pkg, args, resources}
}

// NewGraphSnapshot takes an object graph and produces a resource snapshot from it.  It understands how to name
// resources based on their position within the graph and how to identify and record dependencies.  This function can
// fail dynamically if the input graph did not satisfy the preconditions for resource graphs (like that it is a DAG).
func NewGraphSnapshot(ctx *Context, pkg tokens.PackageName, args core.Args, g *heapstate.ObjectGraph) (Snapshot, error) {
	// First create the monikers, resource objects, and maps that we will use.
	createResources(ctx, g)

	// Next remember the topologically sorted list of resources (in dependency order).  This happens here, rather than
	// lazily on-demand, so that we can hoist errors pertaining to DAG-ness, etc. to the snapshot creation phase.
	resources, err := topsort(ctx, g)
	if err != nil {
		return nil, err
	}

	return NewSnapshot(ctx, pkg, args, resources), nil
}

type snapshot struct {
	ctx       *Context           // the context shared by all operations in this snapshot.
	pkg       tokens.PackageName // the package from which this snapshot came.
	args      core.Args          // the arguments used to compile this package.
	resources []Resource         // the topologically sorted linearized list of resources.
}

func (s *snapshot) Ctx() *Context           { return s.ctx }
func (s *snapshot) Pkg() tokens.PackageName { return s.pkg }
func (s *snapshot) Args() core.Args         { return s.args }
func (s *snapshot) Resources() []Resource   { return s.resources }

func (s *snapshot) ResourceByID(id ID, t tokens.Type) Resource {
	contract.Failf("TODO: not yet implemented")
	return nil
}

func (s *snapshot) ResourceByMoniker(m Moniker) Resource     { return s.ctx.Mks[m] }
func (s *snapshot) ResourceByObject(obj *rt.Object) Resource { return s.ctx.Res[obj] }

// createResources uses a graph to create monikers and resource objects for every resource within.  It
// returns two maps for further use: a map of vertex to its new resource object, and a map of vertex to its moniker.
func createResources(ctx *Context, g *heapstate.ObjectGraph) {
	// First create all of the resource monikers within the graph.
	omks := createMonikers(g)

	// Every moniker entry represents a resource.  Make an object and reverse map entry out of it.
	for o, m := range omks {
		r := NewObjectResource(o, omks)
		ctx.Res[o] = r
		ctx.Mks[m] = r
	}
}

type objectMonikerMap map[*rt.Object]Moniker

// createMonikers walks the graph creates monikers for every one using the algorithm specified in moniker.go.
// In particular, it uses the shortest known graph reachability path to the resource to create a string name.
func createMonikers(g *heapstate.ObjectGraph) objectMonikerMap {
	visited := make(map[graph.Vertex]int)
	mks := make(objectMonikerMap)

	for _, root := range g.Objs() {
		var path []graph.Edge
		createMonikersEdge(root, path, visited, mks)
	}

	return mks
}

// createMonikersEdge inspects a single edge and produces monikers for all resource nodes.  visited keeps track of the
// shortest path distances for vertices we've already seen, and vmks is a map from vertex to its moniker.
func createMonikersEdge(e *heapstate.ObjectEdge, path []graph.Edge, visited map[graph.Vertex]int, mks objectMonikerMap) {
	visit := true          // if we've already visited the shortest path, we won't go further.
	path = append(path, e) // append this edge to the path.

	// If this node is a resource, pay special attention to it.
	v := e.ToObj()
	if IsResourceVertex(v) {
		shortest, exists := visited[v]
		if exists && shortest < len(path) {
			visit = false // if the existing path is shorter, then we can skip it.
		}
		if visit {
			moniker := NewMoniker(v, path)
			// If exists, there will be res and mks entries already.  We will need to patch them up.  Note that, if
			// the path lengths are equal, we pick the moniker that lexicographically comes first.
			obj := v.Obj()
			if exists && shortest == len(path) && moniker > mks[obj] {
				visit = false // the existing moniker sorts before the new one; we can stop.
			} else {
				mks[obj] = moniker
				visited[v] = len(path)
			}
		}
	}

	// For all nodes, keep chasing the out-edges; even for non-resources, we might eventually reach one.
	if visit {
		for _, out := range v.OutObjs() {
			createMonikersEdge(out, path, visited, mks)
		}
	}
}

// topsort actually performs a topological sort on a resource graph.
func topsort(ctx *Context, g graph.Graph) ([]Resource, error) {
	var linear []Resource

	// TODO: we want this to return a *graph*, not a linearized list, so that we can parallelize.

	// Sort the graph output so that it's a DAG; if it's got cycles, this can fail.
	sorted, err := graph.Topsort(g)
	if err != nil {
		return linear, err
	}

	// Now walk the list and prune out anything that isn't a resource.
	for _, v := range sorted {
		ov := v.(*heapstate.ObjectVertex)
		if IsResourceVertex(ov) {
			r := ctx.Res[ov.Obj()]
			contract.Assert(r != nil)
			linear = append(linear, r)
		}
	}

	return linear, nil
}
