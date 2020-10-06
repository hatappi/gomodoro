# Record Pomodoro count with Pixela
Pixela records and tracks your habits or effort. All by API.  
https://pixe.la/  
https://medium.com/@a.know.dev/pixela-a-service-for-generating-github-like-graphs-5867baaa107b

Gomodoro records pomodoro count using Pixela.

## Preparation
You need to prepare Pixela's `token`, `user_name` and `graph_id`.

### token and user_name
https://docs.pixe.la/entry/post-user

e.g. 

```sh
$ curl -X POST \
	https://pixe.la/v1/users \
	-d '{"token":"thisissecret", "username":"hatappi", "agreeTermsOfService":"yes", "notMinor":"yes"}'
```

## graph_id
https://docs.pixe.la/entry/post-graph

e.g.

```sh
$ curl -X POST \
	https://pixe.la/v1/users/hatappi/graphs \
	-H "X-USER-TOKEN:thisissecret" \
	-d '{"id":"gomodoro","name":"gomodoro","unit":"pomodoro","type":"int","color":"shibafu","timezone":"Asia/Tokyo"}'

```

## Usage
add the Pixela config to gomodoro config as below.

```yaml
pixela:
  enable: true
  token: thisissecret
  user_name: hatappi
  graph_id: gomodoro
```
