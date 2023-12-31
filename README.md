# Propagator

A library to assist in solving constraint satisfaction problems using constraint propagation.

See the `examples` folder for applied examples.

## How to use

A constraint satisfaction problem consists of variables that can each take on one of multiple values.
These values are called the variable domain.
Constraints define relations between these variables, allowing to iteratively reduce the variable domains, until either a solution is found, or the problem turns out to be unsolvable, meaning there is no combination of values that satisfy all constraints.

The first step is defining your problem in terms of variables and constraints. Once you have done that, proceed with writing the implementation as per the next sections.

### 1 - Define a new Constraint Satisfaction Problem

Create a new problem instance on which we will define our variables and constraints. This object serves as a builder for the model.

```go
csp := propagator.NewProblem()
```

### 2 - Add your variables

Define and add the variables for which a value need to be selected.

In the instance of a sudoku puzzle, the variables are the cells with their respective states. The domain values are of type `int` (the numbers 1 to 9).
We could define the `Cell` type as such:
```go
type Cell = propagator.Variable[int]
```

Add the variables to the CSP using `AddVariable` or `AddVariableFromValues`, the latter is useful when we know in advance that all our values will have equal probability.
```go
v0 *Cell := propagator.AddVariableFromValues[int](csp, "v0", []int{1,2,3,4,5,6,7,8,9})
v1 *Cell := propagator.AddVariableFromValues[int](csp, "v1", []int{1,2,3,4,5,6,7,8,9})
```

The returned variable can then be used in your constraints.

### 3 - Add your constraints

Define the constraints that apply to these variables.

To define a constraint, implement the `propagator.Constraint` interface.

In a sudoku puzzle one of the constraints can be a House (combination of 9 cells that contain unique values).
```go
type House struct {
	Cells []*Cell
}

func (h House) Scope() []DomainId {
	return propagator.IdsOf(h.Cells)
}

func (h House) Propagate(mutator *propagator.Mutator) {
	/* ... logic omitted ... */
}
```
`Scope` should return a list of `DomainId` that this constraint applies to. Because each `Variable` is a `Domain`, these can be easily extracted.

The implementation of `Propagate` holds the logic that defines the rules of the problem to be solved.
By passing different mutations to the mutator, changes to the domain are defined.
Often, optimizing this implementation can speed up the solution process significantly.

Add instances of your constraints to the CSP by calling `AddConstraint`.
```go
house := House{Cells: []*Cell{v0, v1}}

csp.AddConstraint(house)
```

### 4 - Build the model

Build the model by calling `Model()` on the `Problem`.
```go
model := csp.Model()
```

### 5 - Solve the model

Solve the model using a solver. Additional `SolverOptions` can be passed when creating a new solver.
```go
solver := propagator.NewSolver(/* options */)

solved := solver.Solve(model)

if (!solved) {
    fmt.Printf("no solution!")
} else {
    fmt.Printf("%s", v0.GetFixedValue())
    fmt.Printf("%s", v1.GetFixedValue())
}
```

## Advanced Usage

### Hidden variables

If you want to define variables that should take part in constraint propagation but should not be picked and solved for (for instance when using the domain values as a constraint to the problem), you can mark these variables as hidden.

```
propagator.AddHiddenVariableFromValues[int](csp, "v", []int{1,2,3,4,5,6,7,8,9})
```

### Solving strategies

Using the `SelectDomainsBy*` and `SelectIndicesBy*` solver options, various solving strategies can be configured.

## Resources

Excellent overview of constraint satisfaction problems: http://aima.cs.berkeley.edu/newchap05.pdf

Slideshow with summary of most important points: https://www.cse.unsw.edu.au/~cs3411/18s1/lect/1page/wk08_CSP.pdf