# Templates

Templates are YAML files that define what software is installed when you launch a workspace.

## Using templates

```bash
# List available templates
prism templates

# Details on a specific template
prism templates info python-ml

# Launch a workspace from a template
prism workspace launch python-ml my-project
prism workspace launch r-research stats-project --size L
```

## Built-in templates

| Template | What's included |
|----------|-----------------|
| `python-ml` | Python, PyTorch, TensorFlow, Jupyter |
| `r-research` | R, RStudio Server, Bioconductor |
| `genomics` | GATK, BWA, samtools, STAR |
| `bioinformatics` | Conda, Snakemake, BioPython |
| `deep-learning` | CUDA, cuDNN, GPU-ready PyTorch |
| `data-science` | Pandas, scikit-learn, DuckDB |
| `hpc-base` | MPI, OpenMP, GCC, CMake |

---

## Template YAML format

Templates live in the `templates/` directory and are written in YAML.

### Minimal example

```yaml
name: my-template
description: Python environment with data science tools
base: ubuntu-22.04-server-lts
architecture: x86_64

packages:
  - python3
  - python3-pip
  - jupyter

build_steps:
  - name: Install Python packages
    script: |
      pip3 install numpy pandas scikit-learn matplotlib

validation:
  - name: Check Python
    script: python3 --version
  - name: Check Jupyter
    script: jupyter --version
```

### Full schema

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Unique template name |
| `description` | Yes | Human-readable description |
| `base` | Yes | Base OS image (e.g. `ubuntu-22.04-server-lts`) |
| `architecture` | No | `x86_64` (default) or `arm64` |
| `inherits` | No | Parent template name (see Inheritance below) |
| `package_manager` | No | `apt`, `conda`, or `dnf` |
| `packages` | No | List of packages to install |
| `users` | No | Additional OS users to create |
| `services` | No | Services to start (e.g. `jupyter`, `rstudio`) |
| `ports` | No | Ports to open in the security group |
| `build_steps` | No | Ordered list of setup scripts |
| `validation` | No | Tests run after build to verify the template works |

### Build steps

Each build step runs a shell script:

```yaml
build_steps:
  - name: Install R packages
    script: |
      Rscript -e "install.packages(c('tidyverse', 'ggplot2'), repos='https://cran.r-project.org')"
    timeout: 1800    # seconds; default 600
```

### Validation

Validation runs after build. A non-zero exit code marks the template as broken:

```yaml
validation:
  - name: R is installed
    script: R --version
  - name: tidyverse loads
    script: Rscript -e "library(tidyverse)"
```

---

## Template inheritance

Templates can inherit from a parent. The child adds to (not replaces) the parent's packages, users, services, and ports. The child's `package_manager` replaces the parent's.

```yaml
name: python-ml-gpu
description: python-ml with GPU support
inherits: python-ml

packages:
  - cuda-toolkit-12
  - cudnn9

build_steps:
  - name: Install GPU PyTorch
    script: pip3 install torch --index-url https://download.pytorch.org/whl/cu121
```

Merging rules:
- `packages`, `users`, `services`: **append** (child adds to parent)
- `ports`: **deduplicate**
- `package_manager`: **override** (child replaces parent)
- `build_steps`: **append** (parent steps run first)

---

## Creating a custom template

1. Create a YAML file in `templates/`:
   ```bash
   cp templates/python-ml.yaml templates/my-template.yaml
   ```

2. Edit the file with your changes.

3. Validate the template:
   ```bash
   prism templates validate my-template
   ```

4. Launch a test workspace:
   ```bash
   prism workspace launch my-template test-build
   ```

---

## Tips

- Start from an existing template rather than from scratch — inheritance saves a lot of work.
- Only include packages you actually need; smaller templates launch faster.
- Add validation steps for the tools your users will rely on most.
- Set `timeout` on long build steps (R package installs, compiling from source).
