# pathmatch

PathMatch is a Go library for flexible matching of URL-like paths against defined templates. It allows for extracting named variables from paths and supports wildcards for more complex matching rules.

The library is useful for routing, resource identification, or any scenario where structured path analysis is required. For instance, given a template `/users/{userID}/data` and an input path `/users/123/data`, PathMatch can confirm the match and extract `userID="123"`.

It also provides a `Walker` type for step-by-step path consumption against a sequence of templates.

## What does PathMatch do?

At its core, PathMatch checks if a given path (like a URL) conforms to a predefined template and extracts any dynamic values from it.

Imagine you have a template for user profiles in your web application:

**Template:** `/users/{userID}/profile`

PathMatch can then process incoming paths against this template:

| Incoming Path           | Does it Match? | Extracted Variables  |
| :---------------------- | :------------- | :------------------- |
| `/users/alice/profile`  | ✅ Yes         | `userID` = `"alice"` |
| `/users/12345/profile`  | ✅ Yes         | `userID` = `"12345"` |
| `/users/profile`        | ❌ No          | (None)               |
| `/users/alice/settings` | ❌ No          | (None)               |

This simple mechanism is powerful for:

- **Routing:** Directing `/users/alice/profile` to the user profile handler.
- **Authorization:** Checking if a user has access to a resource defined by a path.
- **Data Extraction:** Getting the `userID` to fetch data from a database.

## Features

- Match concrete paths against templates with literals, wildcards (`*`, `**`), and named variables.
- Extract variables from matched paths.
- Support for sub-templates in variables, allowing multi-segment captures.
- Step-by-step path matching for hierarchical or multi-stage scenarios.
- Templates are parsed into **protocol buffer** (proto) messages and can be stored or reused efficiently.

## Installation

```bash
go get github.com/tsdkv/pathmatch
```

## Usage

The core operation involves matching a concrete path (e.g., `/users/123/profile`) against a parsed `PathTemplate`. A successful match confirms that the path conforms to the template's structure and allows for the extraction of any defined variables.

### Basic Matching

```go
templatePattern := "/users/{userID}/posts/{postID}"
path := "/users/alice/posts/123"
matched, vars, err := pathmatch.CompileAndMatch(templatePattern, path)
// matched == true
// vars == map[string]string{"userID": "alice", "postID": "123"}
```

### Advanced Matching with Parsed Templates

```go
templatePattern := "/items/{category}/{itemID=**}"
tmpl, _ := pathmatch.ParseTemplate(templatePattern)
matched, vars, err := pathmatch.Match(tmpl, "/items/electronics/tv/samsung/qled80")
// matched == true
// vars == map[string]string{"category": "electronics", "itemID": "/tv/samsung/qled80"}
```

### Step-by-Step Traversal with `Walker`

The `Walker` type allows for a more controlled, step-by-step traversal of a concrete path. You initialize a `Walker` with a concrete path and then use its `Step` method with different `PathTemplate`s to consume the path segment by segment. This is useful for navigating hierarchical structures or applying a sequence of rules.

```go
walker := pathmatch.NewWalker("/users/alice/settings/profile/view")

// Initial state
// walker.Depth() == 0
// walker.Remaining() == "/users/alice/settings/profile/view"
// walker.Variables() == map[string]string{}

userTemplate, _ := pathmatch.ParseTemplate("/users/{userID}")
stepVars, ok, _ := walker.Step(userTemplate)
// stepVars == map[string]string{"userID": "alice"}, ok == true
// walker.Depth() == 1
// walker.Remaining() == "/settings/profile/view"
// walker.Variables() == map[string]string{"userID": "alice"}

settingsTemplate, _ := pathmatch.ParseTemplate("/settings/{section}")
stepVars, ok, _ = walker.Step(settingsTemplate)
// stepVars == map[string]string{"section": "profile"}, ok == true
// walker.Depth() == 2
// walker.Remaining() == "/view" (assuming /settings/{section} matched /settings/profile)
// walker.Variables() == map[string]string{"userID": "alice", "section": "profile"}

// Step back
steppedBack := walker.StepBack() // true
// walker.Depth() == 1
// walker.Remaining() == "/settings/profile/view"
// walker.Variables() == map[string]string{"userID": "alice"}

// Try to match another template
actionTemplate, _ := pathmatch.ParseTemplate("/settings/profile/{action}")
stepVars, ok, _ = walker.Step(actionTemplate)
// stepVars == map[string]string{"action": "view"}, ok == true
// walker.Depth() == 2
// walker.Remaining() == ""
// walker.Variables() == map[string]string{"userID": "alice", "action": "view"}

walker.IsComplete() // true

// Reset the walker
walker.Reset()
// walker.Depth() == 0
// walker.Remaining() == "/users/alice/settings/profile/view"
// walker.Variables() == map[string]string{}
```

### Using the `WalkerBuilder`

For more control over the `Walker`'s behavior, such as setting case-insensitive matching, use the `WalkerBuilder`.

```go
// Create a case-insensitive walker
builder := pathmatch.NewWalkerBuilder("/Users/Alice/Settings")
walker := builder.WithCaseIncensitive().Build()

template, _ := pathmatch.ParseTemplate("/users/{id}")
stepVars, ok, _ := walker.Step(template)
// ok is true, because matching is case-insensitive
// stepVars is map[string]string{"id": "Alice"}
```

## Path Template Syntax

Templates must start with a `/`. Path segments are separated by `/`.

1.  **Literals**:

    - Exact string matches for a path segment (e.g., `users`, `config`).
    - Can contain any character except `/`, `*`, `{`, `}`.

2.  **Variables**:

    - Format: `{variableName}`
    - Acts as a placeholder for a single, dynamic path segment.
    - Example: `/users/{userID}` matches `/users/alice` and captures `userID="alice"`.
    - Variable names must follow the same rules as literals (no `/`, `*`, `{`, `}`).

3.  **Single-Segment Wildcard (`*`)**:

    - Matches exactly one path segment.
    - Example: `/files/*/details` matches `/files/image.png/details` and `/files/document.pdf/details`.
    - The value matched by `*` is not captured as a named variable.

4.  **Multi-Segment Wildcard (`\*\*`)**:

    - Matches zero or more consecutive path segments.
    - **Constraint**: Can only appear as the _last segment_ of a path template.
      - Example: `/data/**` matches `/data`, `/data/foo`, and `/data/foo/bar/baz`.
      - Invalid: `/data/**/config`.
    - The value matched by `**` is not captured as a named variable.

5.  **Variables with Sub-Templates**:

    - Syntax: `{variableName=pattern}`.
    - The `pattern` is a sequence of one or more segments, separated by `/`, and can include literals, `*`, or a single `**` at the end.
    - Example: `/files/{path=**}` matches `/files/a/b/c` and captures `path="a/b/c"`.
    - Limitations:
      - `pattern` cannot be empty.
      - Nested variables are not allowed (e.g., `{var={subvar}}` is invalid).
      - If `**` appears in the pattern, it must be the last segment of the entire template.
        - Example: `/files/{rest=**}` is valid.
        - Example: `/files/{rest=**}/extra` is invalid.

## TODO

- [ ] Add custom types for variables, for example:
  - `/users/{id:int}/profile/{section:string}`
  - `/users/{id:uuid}/profile`
- [ ] Support regex patterns for variable matching, e.g., `{id:[0-9]+}`.
- [ ] Fuzz testing to ensure robustness against malformed paths and templates.
- [ ] Secutiry features, such as escaping or sanitizing paths to prevent injection attacks.
  - Add an option to limit the length of the concrete path being processed to prevent attacks with excessively long strings that consume memory.

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
