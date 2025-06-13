# pathmatch

`pathmatch` is a Go package designed for matching URL-like path strings against templates, extracting variables, and supporting wildcards. It's particularly useful for routing, hierarchical configuration access, or implementing systems like Firebase security rules where path patterns define access or behavior.

The package provides functionalities to:

- Parse path templates with variables (e.g., `/users/{userID}`), single-segment wildcards (`*`), and multi-segment wildcards (`**`).
- Match concrete paths against these parsed templates.
- Extract values for variables defined in the templates.
- Perform step-by-step traversal of a concrete path using a sequence of templates with the `Walker` type.
- Utility functions for common path manipulations like splitting, joining, and cleaning.

## Features

- **Path Template Parsing**: Supports literals, named variables (e.g., `{id}`), single-segment wildcards (`*`), and multi-segment wildcards (`**`).
- **Variable Extraction**: Automatically extracts values for named variables during a successful match.
- **Sub-Templates for Variables**: Variables can have their own simple sub-templates (e.g., `{file=*.jpg}`).
- **`Walker` API**: For fine-grained, step-by-step matching and traversal of a concrete path against multiple template segments.
- **Utility Functions**: `Split`, `Join`, `Clean` for robust path string manipulation.
- **High-Level Matching**: `CompileAndMatch` for quick one-off matches.

## Installation

```bash
go get github.com/tsdkv/pathmatch
```

## Core Concepts

### Path Templates

Path templates are strings that define a pattern to match against concrete paths. They can include:

- **Literals**: Exact string segments (e.g., `users`, `config`).
- **Variables**: Placeholders for dynamic segments, enclosed in curly braces (e.g., `{userID}`, `{document_id}`).
- **Single-Segment Wildcard (`*`)**: Matches exactly one path segment.
- **Multi-Segment Wildcard** (`**`): Matches zero or more path segments. It can only appear as the last segment in a template or as the last segment of a variable's sub-template.
- **Variables with Sub-Templates**: A variable can specify a pattern for the segment(s) it should match (e.g., `{image=*.png}` or `{rest=**}`).

**Example Templates:**

- `/users/{userID}/profile`
- `/files/*`
- `/data/{collection=**}`
- `/articles/{year}/{month}/{slug}`

### Matching

The core operation involves matching a concrete path (e.g., `/users/123/profile`) against a parsed `PathTemplate`. A successful match confirms that the path conforms to the template's structure and allows for the extraction of any defined variables.

### Walker

The `Walker` type allows for a more controlled, step-by-step traversal of a concrete path. You initialize a `Walker` with a concrete path and then use its `Step` method with different `PathTemplate`s to consume the path segment by segment. This is useful for navigating hierarchical structures or applying a sequence of rules.

## Usage

### Basic Matching with `CompileAndMatch`

For simple, one-off matching:

```go
package main

import (
	"fmt"
	"log"

	"github.com/tsdkv/pathmatch"
)

func main() {
	templatePattern := "/users/{userID}/posts/{postID}"
	path := "/users/alice/posts/123"

	matched, vars, err := pathmatch.CompileAndMatch(templatePattern, path)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	if matched {
		fmt.Println("Path matched!")
		fmt.Println("Variables:", vars)
		// Output: Variables: map[postID:123 userID:alice]
	} else {
		fmt.Println("Path did not match.")
	}

	// Example with wildcard
	templatePattern2 := "/files/{category}/*"
	path2 := "/files/images/photo.jpg"
	matched, vars, err = pathmatch.CompileAndMatch(templatePattern2, path2)
	if matched {
		fmt.Printf("Path 2 matched! Vars: %v\n", vars)
		// Output: Path 2 matched! Vars: map[category:images] (* is not captured as a named var)
	}
}
```

### Advanced Matching with Parsed Templates

For scenarios where you might reuse a template or need more control:

```go
package main

import (
	"fmt"
	"log"

	"github.com/tsdkv/pathmatch"
)

func main() {
	templatePattern := "/items/{category}/{itemID=**}"
	tmpl, err := pathmatch.ParseTemplate(templatePattern)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	path1 := "/items/electronics/tv/samsung/qled80"
	matched, vars, err := pathmatch.Match(tmpl, path1)
	if err != nil {
		log.Fatalf("Match error: %v", err)
	}
	if matched {
		fmt.Println("Path 1 matched!")
		fmt.Println("Variables:", vars)
		// Output: Variables: map[category:electronics itemID:tv/samsung/qled80]
	}

	path2 := "/items/books" // Will not match fully
	matched, vars, err = pathmatch.Match(tmpl, path2)
	if !matched {
		fmt.Println("Path 2 did not match as expected.")
	}
}
```

### Step-by-Step Traversal with `Walker`

(Assuming the `Walker` API as previously discussed and documented)

```go
package main

import (
	"fmt"
	"log"

	"github.com/tsdkv/pathmatch")

func main() {
	concretePath := "/databases/mydb/documents/users/alice"
	walker := pathmatch.NewWalker(concretePath)

	// Template 1: Match the database part
	dbTemplate, err := pathmatch.ParseTemplate("/databases/{dbName}/documents")
	if err != nil {
		log.Fatal(err)
	}
	stepVars, ok := walker.Step(dbTemplate)
	if !ok {
		log.Fatalf("Failed to match database part. Remaining: %s", walker.Remaining())
	}
	fmt.Printf("Step 1 - Matched: %v, Vars: %v, Remaining: %s, All Vars: %v\n",
		ok, stepVars, walker.Remaining(), walker.Variables())
	// Output: Step 1 - Matched: true, Vars: map[dbName:mydb], Remaining: /users/alice, All Vars: map[dbName:mydb]

	// Template 2: Match the collection and document ID
	docTemplate, err := pathmatch.ParseTemplate("/{collection}/{docID}")
	if err != nil {
		log.Fatal(err)
	}
	stepVars, ok = walker.Step(docTemplate)
	if !ok {
		log.Fatalf("Failed to match document part. Remaining: %s", walker.Remaining())
	}
	fmt.Printf("Step 2 - Matched: %v, Vars: %v, Remaining: %s, All Vars: %v\n",
		ok, stepVars, walker.Remaining(), walker.Variables())
	// Output: Step 2 - Matched: true, Vars: map[collection:users docID:alice], Remaining: , All Vars: map[collection:users dbName:mydb docID:alice]

	fmt.Println("Is complete:", walker.IsComplete()) // Output: Is complete: true

	// Step back
	if walker.StepBack() {
		fmt.Printf("After StepBack - Remaining: %s, All Vars: %v, Depth: %d\n",
			walker.Remaining(), walker.Variables(), walker.Depth())
		// Output: After StepBack - Remaining: /users/alice, All Vars: map[dbName:mydb], Depth: 1
	}
}
```

## Path Template Syntax

Templates must start with a `/`.

1.  **Literals**:

    - Exact string matches, e.g., `/users`, `/files`.
    - Can contain any character except `/`, `*`, `{`, `}`.

2.  **Variables**:

    - Format: `{variableName}`
    - Example: `/users/{userID}` will match `/users/alice` and capture `userID="alice"`.
    - Variable names must be valid literals.

3.  **Single-Segment Wildcard (`*`)**:

    - Matches any single path segment.
    - Example: `/files/*/details` matches `/files/image.png/details` and `/files/document.pdf/details`.
    - The value matched by `*` is not captured as a named variable.

4.  **Multi-Segment Wildcard** (`**`):

    - Matches zero or more path segments.
    - **Constraint**: Can only appear as the _last segment_ of a path template or as the _last segment within a variable's sub-template_.
    - Example: `/data/**` matches `/data`, `/data/foo`, and `/data/foo/bar/baz`.
    - The value matched by `**` is not captured as a named variable.

5.  **Variables with Sub-Templates**:
    - Format: `{variableName=SubTemplate}`
    - The `SubTemplate` can currently consist of:
      - A literal (e.g., `{filename=report.pdf}`).
      - A single `*` (e.g., `{name=*}`). This is equivalent to just `{name}`.
      - A single `**` (e.g., `{remaining_path=**}`). This allows a variable to capture multiple segments.
    - Example: `/archive/{year}/{file=**}` applied to `/archive/2023/reports/annual/summary.pdf` would capture `year="2023"` and `file="reports/annual/summary.pdf"`.
    - **Constraint**: If `**` is used in a sub-template, it must be the only segment in that sub-template (e.g., `{var=**}` is valid, but `{var=prefix/**}` is not currently supported as a sub-template).

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License
