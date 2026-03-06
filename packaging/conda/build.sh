#!/bin/bash

# Copy binaries to the conda environment bin directory
if [[ "$OSTYPE" == "darwin"* ]]; then
  # macOS
  cp prism ${PREFIX}/bin/
  cp prismd ${PREFIX}/bin/
else
  # Linux
  cp prism ${PREFIX}/bin/
  cp prismd ${PREFIX}/bin/
fi

# Make sure binaries are executable
chmod +x ${PREFIX}/bin/prism
chmod +x ${PREFIX}/bin/prismd

# Generate shell completions
mkdir -p ${PREFIX}/etc/bash_completion.d/
${PREFIX}/bin/prism completion bash > ${PREFIX}/etc/bash_completion.d/prism

if [ -d ${PREFIX}/share/zsh/site-functions ]; then
  mkdir -p ${PREFIX}/share/zsh/site-functions/
  ${PREFIX}/bin/prism completion zsh > ${PREFIX}/share/zsh/site-functions/_prism
fi