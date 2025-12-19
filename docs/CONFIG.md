# Configuration Reference

This document describes the YAML configuration format for installers.

## Basic Structure

```yaml
product:
  name: "Application Name"
  version: "1.0.0"
  publisher: "Company Name"
  icon: "path/to/icon.png"  # optional

variables:
  install_dir: "/opt/myapp"
  app_name: "MyApp"

flows:
  install:
    entry: welcome
    steps: []
  
  uninstall:
    entry: confirm
    steps: []
```

## Product Section

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Display name of the product |
| `version` | string | Yes | Version string |
| `publisher` | string | No | Company or author name |
| `icon` | string | No | Path to application icon |

## Variables Section

Define variables that can be used throughout the configuration with `${variable_name}` syntax:

```yaml
variables:
  install_dir: "/opt/myapp"
  bin_dir: "${install_dir}/bin"
  config_dir: "${install_dir}/config"
```

Variables are also set during installation:
- Form fields store their values
- Tasks can modify context variables

## Flows Section

Each flow is a named installation workflow:

```yaml
flows:
  install:
    entry: welcome  # First step ID
    steps:
      - id: welcome
        title: "Welcome"
        screen:
          type: welcome
      # ... more steps
```

### Flow Properties

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `entry` | string | Yes | ID of the first step |
| `steps` | array | Yes | List of step definitions |

## Steps

Each step defines a screen and optional actions:

```yaml
steps:
  - id: welcome
    title: "Welcome to Setup"
    screen:
      type: welcome
      content: "Welcome message here"
    guards:
      - type: mustAccept
        field: license_accepted
    tasks:
      - type: shell
        command: echo "Hello"
```

### Step Properties

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique step identifier |
| `title` | string | Yes | Display title |
| `screen` | object | Yes | Screen configuration |
| `guards` | array | No | Navigation guards |
| `tasks` | array | No | Tasks to execute on this step |

## Screen Types

### Welcome Screen

```yaml
screen:
  type: welcome
  content: "Welcome to the installation wizard."
```

### License Screen

```yaml
screen:
  type: license
  text: |
    MIT License
    
    Copyright (c) 2025...
```

Or load from file:
```yaml
screen:
  type: license
  file: "LICENSE.txt"
```

### Form Screen

```yaml
screen:
  type: form
  fields:
    - id: install_dir
      label: "Installation Directory"
      type: directory
      default: "/opt/myapp"
      required: true
    
    - id: create_shortcut
      label: "Create desktop shortcut"
      type: checkbox
      default: true
    
    - id: port
      label: "Port Number"
      type: text
      default: "8080"
      validation:
        pattern: "^[0-9]+$"
        message: "Must be a number"
```

#### Field Types

| Type | Description |
|------|-------------|
| `text` | Single-line text input |
| `password` | Password input (masked) |
| `directory` | Directory chooser |
| `file` | File chooser |
| `checkbox` | Boolean checkbox |
| `select` | Dropdown selection |

#### Field Properties

| Property | Type | Description |
|----------|------|-------------|
| `id` | string | Variable name to store value |
| `label` | string | Display label |
| `type` | string | Field type |
| `default` | any | Default value |
| `required` | bool | Whether field is required |
| `validation` | object | Validation rules |

### Progress Screen

```yaml
screen:
  type: progress
  message: "Installing..."
```

The progress screen automatically displays task execution progress.

### Summary Screen

```yaml
screen:
  type: summary
  items:
    - label: "Installation Directory"
      value: "${install_dir}"
    - label: "Components"
      value: "Core, Plugins, Documentation"
```

### Finish Screen

```yaml
screen:
  type: finish
  message: "Installation completed successfully!"
  actions:
    - label: "Open README"
      command: "xdg-open ${install_dir}/README.md"
    - label: "Launch Application"
      command: "${install_dir}/bin/myapp"
```

## Guards

Guards control navigation between steps.

### mustAccept

Requires a boolean context variable to be true:

```yaml
guards:
  - type: mustAccept
    field: license_accepted
    message: "You must accept the license to continue"
```

### fieldNotEmpty

Requires a field to have a non-empty value:

```yaml
guards:
  - type: fieldNotEmpty
    field: install_dir
    message: "Installation directory is required"
```

### diskSpace

Requires minimum free disk space:

```yaml
guards:
  - type: diskSpace
    path: "${install_dir}"
    minimum: 1073741824  # 1 GB in bytes
    message: "Insufficient disk space"
```

### expression

Custom expression evaluation:

```yaml
guards:
  - type: expression
    expr: "${port} >= 1024 && ${port} <= 65535"
    message: "Port must be between 1024 and 65535"
```

## Tasks

Tasks are the actual installation operations.

### download

Download a file from URL:

```yaml
tasks:
  - type: download
    url: "https://example.com/package.tar.gz"
    destination: "${temp_dir}/package.tar.gz"
    checksum:
      algorithm: sha256
      value: "abc123..."
```

### unpack

Extract an archive:

```yaml
tasks:
  - type: unpack
    source: "${temp_dir}/package.tar.gz"
    destination: "${install_dir}"
    format: tar.gz  # or "zip"
```

### copy

Copy files or directories:

```yaml
tasks:
  - type: copy
    source: "./files/*"
    destination: "${install_dir}/"
    recursive: true
```

### symlink

Create symbolic links:

```yaml
tasks:
  - type: symlink
    source: "${install_dir}/bin/myapp"
    target: "/usr/local/bin/myapp"
```

### shell

Execute shell commands:

```yaml
tasks:
  - type: shell
    command: "chmod"
    args:
      - "+x"
      - "${install_dir}/bin/myapp"
    workdir: "${install_dir}"
```

Or run a script:
```yaml
tasks:
  - type: shell
    command: "bash"
    args:
      - "-c"
      - |
        echo "Setting up..."
        mkdir -p ${install_dir}/logs
        touch ${install_dir}/logs/app.log
```

### writeConfig

Generate configuration files:

```yaml
tasks:
  - type: writeConfig
    destination: "${install_dir}/config.json"
    format: json
    content:
      app_name: "${app_name}"
      version: "${version}"
      port: 8080
```

Supported formats: `json`, `yaml`, `text`

### desktopEntry

Create .desktop file for Linux desktop integration:

```yaml
tasks:
  - type: desktopEntry
    name: "${app_name}"
    exec: "${install_dir}/bin/myapp"
    icon: "${install_dir}/icons/app.png"
    categories:
      - Utility
      - Application
    terminal: false
```

### removePath

Remove files or directories (for uninstall):

```yaml
tasks:
  - type: removePath
    path: "${install_dir}"
    recursive: true
```

## Complete Example

```yaml
product:
  name: "MyApp"
  version: "2.0.0"
  publisher: "Example Corp"

variables:
  install_dir: "/opt/myapp"
  app_name: "MyApp"

flows:
  install:
    entry: welcome
    steps:
      - id: welcome
        title: "Welcome"
        screen:
          type: welcome
          content: "Welcome to MyApp installer!"

      - id: license
        title: "License Agreement"
        screen:
          type: license
          file: "LICENSE.txt"
        guards:
          - type: mustAccept
            field: license_accepted

      - id: options
        title: "Installation Options"
        screen:
          type: form
          fields:
            - id: install_dir
              label: "Install to"
              type: directory
              default: "/opt/myapp"
            - id: create_shortcut
              label: "Create desktop shortcut"
              type: checkbox
              default: true

      - id: install
        title: "Installing"
        screen:
          type: progress
        tasks:
          - type: shell
            command: mkdir
            args: ["-p", "${install_dir}"]
          
          - type: copy
            source: "./files/*"
            destination: "${install_dir}/"
          
          - type: writeConfig
            destination: "${install_dir}/config.json"
            format: json
            content:
              version: "${version}"
          
          - type: desktopEntry
            name: "${app_name}"
            exec: "${install_dir}/bin/myapp"
            condition: "${create_shortcut}"

      - id: finish
        title: "Complete"
        screen:
          type: finish
          message: "MyApp has been installed successfully!"

  uninstall:
    entry: confirm
    steps:
      - id: confirm
        title: "Confirm Uninstall"
        screen:
          type: summary
          items:
            - label: "Application"
              value: "${app_name}"
            - label: "Location"
              value: "${install_dir}"

      - id: remove
        title: "Uninstalling"
        screen:
          type: progress
        tasks:
          - type: removePath
            path: "${install_dir}"
            recursive: true
          
          - type: shell
            command: rm
            args: ["-f", "~/.local/share/applications/myapp.desktop"]

      - id: done
        title: "Complete"
        screen:
          type: finish
          message: "MyApp has been uninstalled."
```
