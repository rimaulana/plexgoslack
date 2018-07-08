# Plex Go Slack

[![Go Report Card](https://goreportcard.com/badge/github.com/rimaulana/plexgoslack)](https://goreportcard.com/report/github.com/rimaulana/plexgoslack) [![CircleCI](https://img.shields.io/circleci/project/github/rimaulana/plexgoslack.svg)](https://circleci.com/gh/rimaulana/plexgoslack/tree/master) [![codecov](https://codecov.io/gh/rimaulana/plexgoslack/branch/master/graph/badge.svg)](https://codecov.io/gh/rimaulana/plexgoslack) [![codebeat badge](https://codebeat.co/badges/c217e8a8-b808-4b35-aee0-a0705874289d)](https://codebeat.co/projects/github-com-rimaulana-plexgoslack-master) [![Maintainability](https://api.codeclimate.com/v1/badges/4a663411cfea93342333/maintainability)](https://codeclimate.com/github/rimaulana/plexgoslack/maintainability) [![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

Here in my office, we really love to gather our resources by putting together our movies collection into our private Plex server. However, sometimes people didn't know when someone added new movie into our collection and could mistakenly
download the same movie. So we got an Idea to create Slack bot that will inform us if there is a new movie available in our collection by sending a post about it on our Slack general channel and here is how we accomplished it.

![alt text](screenshots/slack-message.png "Slack Message Format")

## Table of Contents
- [Requirements](#requirements)
- [Prerequisites](#prerequisites)
  - [Creating New Slack Webhook](#creating-new-slack-webhook)
  - [Getting TMDb API Key](#getting-tmdb-api-key)
  - [Getting Section Number of Plex Movie Library](#getting-section-number-of-plex-movie-library)
    - [Getting Plex CLI Location](#getting-plex-cli-location)
    - [Setting Plex Environment Variable](#setting-plex-environment-variable)
    - [Getting Section Number](#getting-section-number)
- [Installation](#installation)
  - [Manual Compilation](#manual-compilation)
    - [Getting the Codes](#getting-the-codes)
    - [Compiling the Codes](#compiling-the-codes)
- [Config File](#config-file)
- [Running the Program](#running-the-program)
- [Limitations](#limitations)

## Requirements

* Plex Media Server running on Linux
* Slack Workspace
* Golang version ^1.6 (for manual compilation)
* The Movie Database account  
[back to table of contents](#table-of-contents)

## Prerequisites
### Creating new Slack Webhook

Slack webhook is required in order for this program to be able to send an update to Slack workspace. In order to do that you can add Incoming WebHooks in Slack Custom Integrations. for further information you can read [this reading](https://api.slack.com/incoming-webhooks) provided by Slack. Once done, make sure to get the webhook url  
[back to table of contents](#table-of-contents)

### Getting TMDb API Key

One of the magic of this program is being able to pull the information regarding the movie that has just been added into Plex collection. The iformation that it will get are Movie poster and synopsis. In order to do that, we don't have to reinvent the wheel, all we need to do is using TMDb API to get this information. You can easily follow [this guide](https://developers.themoviedb.org/3/getting-started/introduction) to get it  
[back to table of contents](#table-of-contents)

### Getting Section Number of Plex Movie Library

This one is going to be a little bit challenging since we need to get access to Linux shell that host Plex Media Server.  
[back to table of contents](#table-of-contents)

#### Getting Plex CLI Location

you need to find your Plex CLI location, by default it is located in /usr/lib/plexmediaserver. However, if you couldn't find it in default location, try finding it using the following command.

```bash
sudo find / -name "Plex Media Scanner"
```

then the root path is the path not include the "Plex Media Scanner"  
[back to table of contents](#table-of-contents)

#### Setting Plex Environment Variable

In order for plex cli application to be able to run it will need a specific environment variable named LD_LIBRARY_PATH that I found out by the default it is not set. you can do it by edit /etc/environment file and add the location of your root Plex CLI from the previous step. for example if you found it in the default location

```text
LD_LIBRARY_PATH="/usr/lib/plexmediaserver"
```  
[back to table of contents](#table-of-contents)

#### Getting Section Number

Now that you set up all the required steps, you can run the following command to get section number of your Movie Library

```bash
sudo -u plex -E -H "$LD_LIBRARY_PATH/Plex Media Scanner" --list
```

it will give you output that will look like the following

```text
  1: Movies
  3: Tutorial AWS
  4: Tutorial Data Warehouse
  2: TV Shows
```

from the example above I can say that the section number for my Movie library is 1  
[back to table of contents](#table-of-contents)

## Installation
There are two way to get the binary of the program, first you can get it from our [release page](https://github.com/rimaulana/plexgoslack/releases), or you can build and compile it manually.  
[back to table of contents](#table-of-contents)

### Manual Compilation
#### Getting the Codes

In order to get the code, you can simply run go get command from console

```bash
go get github.com/rimaulana/plexgoslack
```  
[back to table of contents](#table-of-contents)

#### Compiling the Codes

you can change directory into you go workspace and into /src/github.com/rimaulana/plexgoslack

```bash
make release
```

That will yield a file named plexgoslack-vlatest-linux-amd64 under release folder  
[back to table of contents](#table-of-contents)

## Config File

Config file for this program is located in the folder, the name of the file is config.toml and it has the following structure

```toml
# Is the url of your plex media server page for example https://app.plex.tv
plex_url = "link to your plex server page"

# The API Key you get on step Getting TMDb API Key
[tmdb]
api_key = "The movie databse API Key"

# Is an array contains the webhook URL to your slack incoming webhook integration. it can be multiple webhooks
[slack]
webhooks = ["slack_webhook_1","slack_webhook_2"]


# This is where you put information on each library you want to watch if there are changes. It can be multiple libraris but you need to see the limitations
[plex]
[plex.movies] # the naming after plex. is up to you
root = "/path/to/movie" #path where you keep you movie2 collection
section = 1 #int respresent plex section number

[plex.movies2] # the naming after plex. is up to you
root = "/path/to/movie2" #path where you keep you movie2 collection
section = 2 #int respresent plex section number
```  
[back to table of contents](#table-of-contents)

## Running the Program

In order to run the program, you need to run the binary file plexgoslack you downloaded from our [release page](https://github.com/rimaulana/plexgoslack/releases) or generated on step [Compiling the Codes](#compiling-the-codes). config.toml file needs to be on the same folder as plexgoslack binary. Before running the program you need to add execute permission on it by running

```bash
sudo chmod +x plexgoslack-version-linux-amd64
```

once done, run the code with sudo permission

```bash
sudo ./plexgoslack-version-linux-amd64
```  
[back to table of contents](#table-of-contents)

## Limitations

So far this program can only monitor Plex Movie Library type. It only read the name of the parent folder of each movie item in the folder. The pattern that the file watcher looking for [Slack movie Folder Nesting naming standard](https://support.plex.tv/hc/en-us/articles/200381023-Naming-Movie-files), if it doesn't match the regex, it will not be considered as a new movie item and will not be updated on Plex and on Slack  
[back to table of contents](#table-of-contents)