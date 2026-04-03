# Quick Start

Get a research workstation running in under 5 minutes.

## 1. Install

=== "macOS (Homebrew)"
    ```bash
    brew install scttfrdmn/tap/prism
    ```

=== "Windows (Scoop)"
    ```powershell
    scoop bucket add scttfrdmn https://github.com/scttfrdmn/scoop-bucket
    scoop install prism
    ```

=== "Linux"
    Download the latest tarball from the [releases page](https://github.com/scttfrdmn/prism/releases/latest), extract, and add to your PATH.

## 2. Connect AWS credentials

If you haven't configured the AWS CLI yet:

```bash
aws configure
```

Then add a Prism profile pointing to those credentials:

```bash
prism profile add
```

## 3. Launch a workspace

```bash
# See what's available
prism templates

# Launch a Python + Jupyter environment
prism workspace launch python-ml my-project

# Check it's running (takes ~2 minutes)
prism workspace list
```

## 4. Connect

```bash
prism workspace connect my-project
```

This prints the SSH command and any web service URLs (Jupyter, RStudio, etc.).

## 5. Stop when done

```bash
prism workspace stop my-project
```

Stopped workspaces don't cost anything. Resume later with `prism workspace start my-project`, or delete permanently with `prism workspace delete my-project`.

---

## Common templates

| Template | What's included |
|----------|-----------------|
| `python-ml` | Python, PyTorch, TensorFlow, Jupyter |
| `r-research` | R, RStudio Server, Bioconductor |
| `genomics` | GATK, BWA, samtools, STAR |
| `bioinformatics` | Conda, Snakemake, BioPython |
| `deep-learning` | CUDA, cuDNN, GPU-ready PyTorch |
| `data-science` | Pandas, scikit-learn, DuckDB |
| `hpc-base` | MPI, OpenMP, GCC, CMake |

Run `prism templates info <name>` for details on any template.

---

## What's next

- **[Installation & Setup](GETTING_STARTED.md)** — full install options, IAM setup, first-run wizard
- **[CLI Reference](CLI_REFERENCE.md)** — all commands
- **[GUI Guide](GUI_USER_GUIDE.md)** — desktop app
