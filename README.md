# gomodoro
gomodoro is Pomodoro Technique by Go.  
This record working time in [toggl](https://toggl.com/).

## Installation

```
go get github.com/hatappi/gomodoro
```

or  

https://github.com/hatappi/gomodoro/releases

## Usage

### start command

```bash
$ gomodoro start
```

**1.Select task**  
※ add task when there is no task.  
The cursor moves down by pressing `j`, and up by pressing `k`.  
select `Enter`.

**2.Repeat working and break**  
When remaining time runs out, please press any key.  
The next step begins.
At this time only working time is recorded in [toggl](https://toggl.com/).

#### options
- `--long-break-sec value, -l value`:  long break (s) (default: 900)
- `--short-break-sec value, -s value`: short break (s) (default: 300)
- `--work-sec value, -w value`: work (s) (default: 1500)

### remain command

display remain time.

````bash
$ gomodoro remain
````


#### options
- `--ignore-error, -i`: ignore errors

## global options
- `--conf-path value, -c value`: gomodoro config path (default: "~/.gomodoro/config.toml")
- `--app-dir value, -a value`: application directory (default: "~/.gomodoro")
- `--socket-path value, -s value`: gomodoro socket path (default: "/tmp/gomodoro.sock")

## Config

```
[Toggl]
# Toggl Api token ref: https://toggl.com/app/profile
ApiToken = "xxx"
# ProjectId ref: https://toggl.com/app/projects/xxxxx/edit/yyyyy
PID=yyyyy
```

## Support
- OSX
- Linux

