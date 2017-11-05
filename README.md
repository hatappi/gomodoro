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

```
$ gomodoro
```

**1.Select task**  
※ add task when there is no task.  
The cursor moves up by pressing `j`, and down by pressing `k`.  
select `Enter`.

**2.Repeat working and break**  
When remaining time runs out, please press any key.  
The next step begins.
At this time only working time is recorded in [toggl](https://toggl.com/).

## Option
- config: gomodoro config path. default `~/.gomodoro/config.toml`

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

