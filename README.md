# Gomodoro
Gomodoro is Pomodoro Technique by Go.  
The working time can be automatically recorded in [toggl](https://toggl.com/).

[![CI](https://github.com/hatappi/gomodoro/actions/workflows/ci.yaml/badge.svg)](https://github.com/hatappi/gomodoro/actions/workflows/ci.yaml)
[![release](https://github.com/hatappi/gomodoro/actions/workflows/release.yaml/badge.svg)](https://github.com/hatappi/gomodoro/actions/workflows/release.yaml)

## Installation

```sh
go get github.com/hatappi/gomodoro
```

or  

https://github.com/hatappi/gomodoro/releases

## Usage
first of all, run `gomodoro init`.  

```
$ gomodoro init
success to create config file. (/Users/user/.gomodoro/config.yaml)
```

if you wanna record working time to [toggl](https://toggl.com/), please edit config file.

### start command

```bash
$ gomodoro start
```

**1.Select task**  
※ add task when there is no task.  
The cursor moves down by pressing `j` or down key, and up by pressing `k` or up key.  
select `Enter`.

**2.Repeat working and break**  
When remaining time runs out, please press Enter. The next step begins.  
At this time only working time is recorded in [toggl](https://toggl.com/) if you setting.

### remain command

you can see remain time if gomodoro already running.

````bash
$ gomodoro remain
````
