# pathmatch

**PathMatch** is a Go library for matching and extracting variables from URL-like path templates. It supports features such as wildcards, named variables, and sub-templates for flexible path matching.

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

### Matching

The core operation involves matching a concrete path (e.g., `/users/123/profile`) against a parsed `PathTemplate`. A successful match confirms that the path conforms to the template's structure and allows for the extraction of any defined variables.

### Walker

The `Walker` type allows for a more controlled, step-by-step traversal of a concrete path. You initialize a `Walker` with a concrete path and then use its `Step` method with different `PathTemplate`s to consume the path segment by segment. This is useful for navigating hierarchical structures or applying a sequence of rules.

## Usage

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
// vars == map[string]string{"category": "electronics", "itemID": "tv/samsung/qled80"}
```

### Step-by-Step Traversal with Walker

```go
walker := pathmatch.NewWalker("/users/alice/settings/profile")

userTemplate, _ := pathmatch.ParseTemplate("/users/{userID}")
stepVars, ok, _ := walker.Step(userTemplate)
// stepVars == map[string]string{"userID": "alice"}, ok == true

settingsTemplate, _ := pathmatch.ParseTemplate("/settings/{section}")
stepVars, ok, _ = walker.Step(settingsTemplate)
// stepVars == map[string]string{"section": "profile"}, ok == true

walker.IsComplete() // true
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

- [ ] **Optional segments** (e.g., `/foo/bar?/baz`, `/foo/{var?}/baz`).
- [ ] Options like case sensitivity, strict matching, etc.
- [ ] Add custom types for variables, for example:
  - `/users/{id:int}/profile/{section:string}`
  - `/users/{id:uuid}/profile`
- [ ] Support regex patterns for variable matching, e.g., `{id:[0-9]+}`.
- [ ] Implement more complex sub-template matching, such as allowing `**` in the middle of a sub-template (e.g., `{var=prefix/**/suffix}`).
- [ ] Fuzz testing to ensure robustness against malformed paths and templates.
- [ ] Secutiry features, such as escaping or sanitizing paths to prevent injection attacks.

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
