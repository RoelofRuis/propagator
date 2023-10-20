# Propagator

A library to assist in solving constraint satisfaction problems using constraint propagation.

See the `examples` folder for applied examples.

## How to use

A constraint satisfaction problem consists of variables that can each take on one of multiple values.
These values are called the variable domain.
Constraints define relations between these variables, allowing to iteratively reduce the variable domains, until either a solution is found, or the problem turns out to be unsolvable, meaning there is no combination of values that satisfy all constraints.

Once you have modeled your problem in terms of variables and constraints, proceed to the next steps.

### 1 - Define your variables

Define the variables for which a value need to be selected.

In the instance of a sudoku puzzle, the variables are the cells with their respective states. The state values are of type `int` (the numbers 1 to 9)
```go
type Cell = propagator.Variable[int]
```

We could instantiate the variable by hand or use one of the helper functions provided. If, for instance, the set of values is known and all have equal probability, we can use:
```go
var variable *Cell := propagator.NewVariableFromValues[int]("name", []int{1,2,3,4,5,6,7,8,9})
```

### 2 - Define your constraints

Define the constraints that apply to these variables.

To define a constraint, implement the `propagator.Constraint` interface.

In a sudoku puzzle one of the constraints can be a House (combination of 9 cells that contain unique values).
```go
type House struct {
	Cells []*Cell
}

func (h House) Scope() []*propagator.Domain {
	return propagator.DomainsOf(h.Cells)
}

func (h House) Propagate(mutator *propagator.Mutator) {
	/* ... logic omitted ... */
}
```
`Scope` should return a list of domains that this constraint applies to. As Variables are built on top of domains, these can be easily extracted.

The implementation of `Propagate` holds the most important logic. By passing different mutations to the mutator, changes to the domain are defined.
Often, optimizing this implementation can speed up the solution process significantly.

### 3 - Build a model

Build a model using your variables and constraints.
```go
builder := propagator.BuildModel()

builder.AddDomain(variable.Domain)
builder.AddConstraint(House{})

model := builder.Build()
```

### 4 - Solve the model

Solve the model using a solver. Additional `SolverOptions` can be passed when creating a new solver.
```go
solver := propagator.NewSolver(/* options */)

solved := solver.Solve(model)

if (!solved) {
    fmt.Printf("no solution!")
} else {
    fmt.Printf("%s", variable.GetFixedValue())
}
```

## Resources

Excellent overview of constraint satisfaction problems: http://aima.cs.berkeley.edu/newchap05.pdf