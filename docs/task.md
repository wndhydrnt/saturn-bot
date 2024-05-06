# Task

## actions

[json-path:../pkg/task/schema/task.schema.json:$.properties.actions.description]

### fileCreate

[json-path:../pkg/task/schema/task.schema.json:$.properties.actions.properties.fileCreate.description]

#### Properties

| Property | Description | Type | Default | Required |
| --- | --- | --- | --- | --- |
| **content** | [json-path:../pkg/task/schema/task.schema.json:$.properties.actions.properties.fileCreate.items.properties.content.description] | string | `""` | No |
| **contentFile** | [json-path:../pkg/task/schema/task.schema.json:$.properties.actions.properties.fileCreate.items.properties.contentFile.description] | string | `""` | No |
| **mode** | [json-path:../pkg/task/schema/task.schema.json:$.properties.actions.properties.fileCreate.items.properties.mode.description] | integer | `644` | No |
| **overwrite** | [json-path:../pkg/task/schema/task.schema.json:$.properties.actions.properties.fileCreate.items.properties.overwrite.description] | boolean | `true` | No |
| **path** | [json-path:../pkg/task/schema/task.schema.json:$.properties.actions.properties.fileCreate.items.properties.path.description] | string | N/A | Yes |

#### Examples

```yaml
# Create the file hello-world.txt with content "Hello World"
# at the root of the repository.
actions:
  fileCreate:
    - content: "Hello World"
      path: "hello-world.txt"
```

```yaml
# Create the file hello-world.txt with content "Hello World"
# at the root of the repository.
# Do nothing if the file already exists.
actions:
  fileCreate:
    - content: "Hello World"
      path: "hello-world.txt"
      overwrite: false
```

```yaml
# Create the file hello-world.txt at the root of the repository.
# Read the content from the file content.txt.
# The path of content.txt is relative to the path of the task.
actions:
  fileCreate:
    - contentFile: "content.txt"
      path: "hello-world.txt"
```

### fileDelete

[json-path:../pkg/task/schema/task.schema.json:$.properties.actions.properties.fileDelete.description]

#### Properties

| Property | Description | Type | Default | Required |
| --- | --- | --- | --- | --- |
| **path** | [json-path:../pkg/task/schema/task.schema.json:$.properties.actions.properties.fileDelete.items.properties.path.description] | string | N/A | Yes |

#### Examples

```yaml
# Delete the file hello-world.txt in the root of the repository.
actions:
  fileDelete:
    - path: "hello-world.txt"
```

### lineDelete

[json-path:../pkg/task/schema/task.schema.json:$.properties.actions.properties.lineDelete.description]

#### Properties

| Property | Description | Type | Default | Required |
| --- | --- | --- | --- | --- |
| **line** | [json-path:../pkg/task/schema/task.schema.json:$.properties.actions.properties.lineDelete.items.properties.line.description] | string | N/A | Yes |
| **path** | [json-path:../pkg/task/schema/task.schema.json:$.properties.actions.properties.lineDelete.items.properties.path.description] | string | N/A | Yes |

#### Examples

```yaml
# Delete every line that equals "Hello World" in the file hello-world.txt.
actions:
  lineDelete:
    - line: "Hello World"
      path: "hello-world.txt"
```

```yaml
# Delete every line that starts with "Hello" in every .txt file.
actions:
  lineDelete:
    - line: "Hello.+"
      path: "*.txt"
```

### lineInsert

[json-path:../pkg/task/schema/task.schema.json:$.properties.actions.properties.lineInsert.description]

#### Properties

| Property | Description | Type | Default | Required |
| --- | --- | --- | --- | --- |
| **insertAt** | [json-path:../pkg/task/schema/task.schema.json:$.properties.actions.properties.lineInsert.items.properties.insertAt.description] | string | `"EOF"` | No |
| **line** | [json-path:../pkg/task/schema/task.schema.json:$.properties.actions.properties.lineInsert.items.properties.line.description] | string | N/A | Yes |
| **path** | [json-path:../pkg/task/schema/task.schema.json:$.properties.actions.properties.lineInsert.items.properties.path.description] | string | N/A | Yes |

#### Examples

```yaml
# Insert the line at the end of the file hello-world.txt.
actions:
  lineInsert:
    - line: "Hello Another World"
      path: "hello-world.txt"
```

```yaml
# Insert the line at the end of each .txt file.
actions:
  lineInsert:
    - line: "Hello Another World"
      path: "*.txt"
```

```yaml
# Insert the line at the beginning of the file hello-world.txt.
actions:
  lineInsert:
    - insertAt: "BOF"
      line: "Hello Another World"
      path: "hello-world.txt"
```

### lineReplace

[json-path:../pkg/task/schema/task.schema.json:$.properties.actions.properties.lineReplace.description]

#### Properties

| Property | Description | Type | Default | Required |
| --- | --- | --- | --- | --- |
| **line** | [json-path:../pkg/task/schema/task.schema.json:$.properties.actions.properties.lineReplace.items.properties.line.description] | string | N/A | Yes |
| **path** | [json-path:../pkg/task/schema/task.schema.json:$.properties.actions.properties.lineReplace.items.properties.path.description] | string | N/A | Yes |
| **search** | [json-path:../pkg/task/schema/task.schema.json:$.properties.actions.properties.lineReplace.items.properties.search.description] | string | N/A | Yes |

#### Examples

```yaml
# Replace line "Hello World" with "World Hello" in file hello-world.txt
actions:
  lineReplace:
    - line: "World Hello"
      path: "hello-world.txt"
      search: "Hello World"
```

## autoMerge

[json-path:../pkg/task/schema/task.schema.json:$.properties.autoMerge.description]

Examples

```yaml
# Enable auto-merge behavior
autoMerge: true
```

```yaml
# Disable auto-merge behavior
autoMerge: false
```

## autoMergeAfter

[json-path:../pkg/task/schema/task.schema.json:$.properties.autoMergeAfter.description]

Examples

```yaml
# Merge pull request automatically.
autoMerge: true
```

```yaml
# Don't merge pull request automatically.
autoMerge: false
```

## branchName

[json-path:../pkg/task/schema/task.schema.json:$.properties.branchName.description]

Examples

```yaml
# Set a custom name.
branchName: "feature/hello-world"
```

## changeLimit

[json-path:../pkg/task/schema/task.schema.json:$.properties.changeLimit.description]

## commitMessage

[json-path:../pkg/task/schema/task.schema.json:$.properties.commitMessage.description]

## createOnly

[json-path:../pkg/task/schema/task.schema.json:$.properties.createOnly.description]

## disabled

[json-path:../pkg/task/schema/task.schema.json:$.properties.disabled.description]

## filters

### repositoryName

### file

### fileContent

## keepBranchAfterMerge

[json-path:../pkg/task/schema/task.schema.json:$.properties.keepBranchAfterMerge.description]

## labels

[json-path:../pkg/task/schema/task.schema.json:$.properties.labels.description]

## mergeOnce

[json-path:../pkg/task/schema/task.schema.json:$.properties.mergeOnce.description]

## name

[json-path:../pkg/task/schema/task.schema.json:$.properties.name.description]

## plugins

## prBody

[json-path:../pkg/task/schema/task.schema.json:$.properties.prBody.description]

## prTitle

[json-path:../pkg/task/schema/task.schema.json:$.properties.prTitle.description]
