# Rotation Scheduler
A GitHub Action for generating a rotation schedule.

## General Usage

Build the binary:

```bash
go build rotation.go
```

Generate a fresh schedule:

```bash
$ rotation schedule generate --start 2020-03-01 --stop 2020-04-01 --users abc,lmn,xyz
shifts:
- startDate: Sun 01 Mar 2020
  user: abc
- startDate: Sun 08 Mar 2020
  user: lmn
- startDate: Sun 15 Mar 2020
  user: xyz
- startDate: Sun 22 Mar 2020
  stopDate: Sat 28 Mar 2020
  user: abc
```

Write the schedule to a file automatically:
```bash
$ rotation schedule generate --start 2020-03-01 --stop 2020-04-01 --users abc,lmn,xyz rotation-schedule.yaml
```

Periodically extend the schedule, optionally changing the users:
```bash
$ rotation schedule extend --schedule rotation-schedule.yaml --stop 2020-05-01 --users abc,lmn,xyz,123 rotation-schedule.yaml
$ cat rotation-schedule.yaml
shifts:
- startDate: Sun 01 Mar 2020
  user: abc
- startDate: Sun 08 Mar 2020
  user: lmn
- startDate: Sun 15 Mar 2020
  user: xyz
- startDate: Sun 22 Mar 2020
  user: abc
- startDate: Sun 29 Mar 2020
  user: lmn
- startDate: Sun 05 Apr 2020
  user: xyz
- startDate: Sun 12 Apr 2020
  user: "123"
- startDate: Sun 19 Apr 2020
  stopDate: Sat 25 Apr 2020
  user: abc
```

Use `--prune` flag to remove past shifts and reschedule shifts when users drop out of rotation:
```bash
$ date '+%Y-%m-%d' # pruning works from the current date. Only whole shifts are removed.
2020-04-09

$ rotation schedule extend --prune --schedule rotation-schedule.yaml --stop 2020-05-01 --users lmn,xyz,123 rotation-schedule.yaml
$ cat rotation-schedule.yaml
shifts:
- startDate: Sun 05 Apr 2020
  user: xyz
- startDate: Sun 12 Apr 2020
  user: "123"
- startDate: Sun 19 Apr 2020
  stopDate: Sat 25 Apr 2020
  user: lmn
```

Includes GitHub Teams integration:
```bash
$ rotation schedule generate --start 2020-03-01 --stop 2020-04-01 --github spinnaker,build-cops,$GITHUB_TOKEN
shifts:
- startDate: Sun 01 Mar 2020
  user: ajordens
- startDate: Sun 08 Mar 2020
  user: cfieber
- startDate: Sun 15 Mar 2020
  user: ethanfrogers
- startDate: Sun 22 Mar 2020
  stopDate: Sat 28 Mar 2020
  user: ezimanyi
```

Sync schedule to Google Calendar. If `user` fields are email addresses, the email address is added as an event attendee.
The `--calendarID` **must** be a G Suite user account dedicated for this purpose, as all existing events are deleted 
during a sync!

The `--jsonKey` is a Google Cloud Platform service account with
[domain-wide delegation](https://developers.google.com/admin-sdk/directory/v1/guides/delegation).
```bash
$ JSON_KEY=$(cat rotation-scheduler.json | base64 -w 0)
$ rotation calendar sync --calendarID build-cop@spinnaker.io --jsonKey $JSON_KEY rotation-schedule.yaml
```
