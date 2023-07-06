# Deploifai CLI

This is the CLI for the Deploifai platform.

## Installation

### Apt (Debian/Ubuntu)

Install using `apt`:

```shell
# install curl, ca-certificates, gnupg
sudo apt install -y curl ca-certificates gnupg && gpg --list-keys

# download gpg public key
curl -fsSL https://packages.deploif.ai/apt/pubkey | sudo gpg --dearmor -o /usr/share/keyrings/deploifai.gpg

# add repo to sources.list
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/deploifai.gpg] https://packages.deploif.ai/apt stable main" | sudo tee -a /etc/apt/sources.list.d/deploifai.list

# update apt
sudo apt update

# install
sudo apt install deploifai
```

Or just run the following script:

```shell
curl -L https://packages.deploif.ai/apt/scripts/cli-install.sh | bash
```

### Homebrew (macOS/Linux)

Install using homebrew:

```shell
# add tap
brew tap deploifai/deploifai

# install
brew install deploifai
```

### Scoop (Windows)

Install using scoop:

```shell
# add bucket
scoop bucket add deploifai

# install
scoop install deploifai
```

## Usage

Login with a personal access token generated from the Deploifai Dashboard:

```shell
deploifai auth login
```

Other useful commands:

```shell
deploifai version
deploifai help
```

## Documentation

For more information, please see the [documentation](https://docs.deploif.ai/cli/commands/quick-start).
