#!/bin/bash
set -e
echo "DEPLOY_DOCS=$DEPLOY_DOCS"
if [[ $DEPLOY_DOCS != true ]]; then
    echo "Not building docs"
    exit 0
fi

echo "Building docs"

# Install mkdocs
virtualenv venv
venv/bin/pip install mkdocs

# Decrypt and load the ssh key
openssl aes-256-cbc -K $encrypted_5704967818cd_key -iv $encrypted_5704967818cd_iv -in docKey.enc -out docKey -d
chmod 0600 docKey
eval `ssh-agent -s`
ssh-add docKey


# Push up to gh-pages
# --force is required otherwise it will fail to push up
venv/bin/mkdocs gh-deploy --remote-name git@github.com:Netflix/chaosmonkey.git --force
