# Slack Orb [![CircleCI Build Status](https://circleci.com/gh/CircleCI-Public/slack-orb.svg?style=shield "CircleCI Build Status")](https://circleci.com/gh/CircleCI-Public/slack-orb) [![CircleCI Orb Version](https://img.shields.io/badge/endpoint.svg?url=https://badges.circleci.io/orb/circleci/slack)](https://circleci.com/orbs/registry/orb/circleci/slack) [![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/CircleCI-Public/slack-orb/master/LICENSE) [![CircleCI Community](https://img.shields.io/badge/community-CircleCI%20Discuss-343434.svg)](https://discuss.circleci.com/c/ecosystem/orbs)

**Easily integrate custom [Slack](https://slack.com/ "Slack") notifications into your [CircleCI](https://circleci.com/ "CircleCI") projects. Create custom alert messages for any job or receive status updates.**

Learn more about [Orbs](https://circleci.com/docs/2.0/using-orbs/ "Using Orbs").

## Usage
Example config:

```yaml
version: 2.1

orbs:
  slack: circleci/slack@x.y.z/*

jobs:
  build:
    docker:
      - image: <docker image>
    steps:
      - slack/<command>
```

`slack@1.0.0` from the `circleci` namespace is imported into the config.yml as `slack` and can then be referenced as such in any job or workflow.

## Commands

### Approval
Send a notification that a manual approval job is ready

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `color` | `string` | '#3AA3E3' | Hex color value for notification attachment color. |
| `mentions` | `string` | '' | A comma separated list of user IDs. No spaces. |
| `message` | `string` | A workflow in CircleCI is awaiting your approval. | Enter custom message. |
| `url` | `string` | 'https://circleci.com/workflow-run/${CIRCLE_WORKFLOW_ID}' | The URL to link back to. |
| `webhook` | `string` | '${SLACK_WEBHOOK}' | Enter either your Webhook value or use the CircleCI UI to add your token under the 'SLACK_WEBHOOK' env var |

Example:

```yaml
version: 2.1

orbs:
  slack: circleci/slack@x.y.z/*

jobs:
    docker:
      - image: <docker image>
    steps:
      - slack/approval:
          message: "This is a custom approval message" # Optional: Enter your own message
          mentions: "USERID1,USERID2," # Optional: Enter the Slack IDs of any user or group (sub_team) to be mentioned
          color: "#42e2f4" # Optional: Assign custom colors for each approval message
          webhook: "webhook" # Optional: Enter a specific webhook here or the default will use $SLACK_WEBHOOK
```

### Notify
Notify a slack channel with a custom message at any point in a job with this custom step.

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `webhook` | `string` | ${SLACK_WEBHOOK} | Either enter your webhook value or use the CircleCI UI to add your token under the `SLACK_WEBHOOK` environment variable |
| `message` | `string` | Your job on CircleCI has completed. | Enter a custom message to send to your Slack channel |
| `mentions` | `string` | `false` | Comma-separated list of Slack User or Group (SubTeam) IDs (e.g., "USER1,USER2,USER3"). _**Note:** these are Slack User IDs, not usernames. The user ID can be found on the user's profile. Look below for information on obtaining Group ID. For `here`, `channel` or `everyone` just write them._ |
| `color` | `string` | #333333 |  Hex color value for notification attachment color |
| `author_name` | `string` |  | Optional author name property for the [Slack message attachment] |
| `author_link` | `string` |  | Optional author link property for the [Slack message attachment] |
| `title` | `string` |  | Optional title property for the [Slack message attachment] |
| `title_link` | `string` |  | Optional title link property for the [Slack message attachment] |
| `footer` | `string` |  | Optional footer property for the [Slack message attachment] |
| `ts` | `string` |  | Optional timestamp property for the [Slack message attachment] |
| `include_project_field` | `boolean` | `true` | Condition to check if it is necessary to include the _Project_ field in the message |
| `include_job_number_field` | `boolean` | `true` | Whether or not to include the _Job Number_ field in the message |
| `include_visit_job_action` | `boolean` | `true` | Whether or not to include the _Visit Job_ action in the message |
| `channel` | `string` | | ID of channel if set, overrides webhook's default channel setting |

[Slack message attachment]: https://api.slack.com/docs/message-attachments

Example:

```yaml
version: 2.1

orbs:
  slack: circleci/slack@x.y.z/*

jobs:
  build:
    docker:
      - image: <docker image>
    steps:
      - slack/notify:
          message: "This is a custom message notification" # Optional: Enter your own message
          mentions: "USERID1,USERID2," # Optional: Enter the Slack IDs of any user or group (sub_team) to be mentioned
          color: "#42e2f4" # Optional: Assign custom colors for each notification
          webhook: "webhook" # Optional: Enter a specific webhook here or the default will use $SLACK_WEBHOOK
```

![Custom Message Example](/img/notifyMessage.PNG)

Refer to Slack's [Basic message formatting](https://api.slack.com/docs/message-formatting) documentation for guidance on formatting notification messages.

### Status
Send a status alert at the end of a job based on success or failure. This must be the last step in a job.

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `webhook` | `string` | ${SLACK_WEBHOOK} | Either enter your webhook value or use the CircleCI UI to add your token under the `SLACK_WEBHOOK` environment variable |
| `success_message` | `string` | :tada: A $CIRCLE_JOB job has succeeded! $SLACK_MENTIONS | Enter your custom message to send to your Slack channel |
| `failure_message` | `string` | :red_circle: A $CIRCLE_JOB job has failed! $SLACK_MENTIONS | Enter your custom message to send to your Slack channel |
| `mentions` | `string` |  | Comma-separated list of Slack User or Group (SubTeam) IDs (e.g., "USER1,USER2,USER3"). _**Note:** these are Slack User IDs, not usernames. The user ID can be found on the user's profile. Look below for information on obtaining Group ID._ |
| `fail_only` | `boolean` | `false` | If set to `true`, successful jobs will _not_ send alerts |
| `only_for_branches` | `string` |  | If set, a comma-separated list of branches, for which to send notifications |
| `include_project_field` | `boolean` | `true` | Whether or not to include the _Project_ field in the message |
| `include_job_number_field` | `boolean` | `true` | Whether or not to include the _Job Number_ field in the message |
| `include_visit_job_action` | `boolean` | `true` | Whether or not to include the _Visit Job_ action in the message |
| `channel` | `string` | | ID of channel if set, overrides webhook's default channel setting |

Example:

```yaml
version: 2.1

orbs:
  slack: circleci/slack@x.y.z/*

jobs:
  build:
    docker:
      - image: <docker image>
    steps:
      # With fail_only set to `true`, no alert will be sent in this example. Change the exit status on the next line to produce an error.
      - run: exit 0

      - slack/status:
          mentions: "USERID1,USERID2" # Optional: Enter the Slack IDs of any user or group (sub_team) to be mentioned
          fail_only: true # Optional: if set to `true` then only failure messages will occur.
          webhook: "webhook" # Optional: Enter a specific webhook here or the default will use $SLACK_WEBHOOK
          only_for_branches: "master" # Optional: If set, a specific branch for which status updates will be sent. In this case, only for pushes to master branch.
```

![Status Success Example](/img/statusSuccess.PNG)
![Status Fail Example](/img/statusFail.PNG)

## Jobs

### approval-notification
Send an approval notification message

Example:

```yaml
version: 2.1

orbs:
  slack: circleci/slack@x.y.z/*

jobs:
  - slack/approval-notification:
      color: "#aa7fcd" # Optional: Enter your own message
      message: "Deployment pending approval" # Optional: Custom approval message
```

## Dependencies / Requirements

### Bash Shell
Because these scripts use bash-specific features, `Bash` is required.
`Bash` is the default shell used on CircleCI and the Orb will be compatible with most images.
If using an `Alpine` base image, you will need to call `apk add bash` before calling this Orb,
or create a derivative base image that calls `RUN apk add bash`.
If `Bash` is not available, an error message will be logged and the task will fail.

### cURL
cURL is used to post the Webhook data and must be installed in the container to function properly.

## Help

### How to get your Slack Webhook
Full instructions can be found at Slack: https://api.slack.com/incoming-webhooks

1. [Create Slack App](https://api.slack.com/docs/slack-button#register_your_slack_app). This will also be the name of the "user" that posts alerts to Slack. You'll be asked for which Workspace this app belongs to.
2. In the settings for the app, enable `Incoming Webhooks`.
3. In the left hand panel of your Slack app settings, under `Features` click `Incoming Webhooks`.
4. Click `Add New Webhook to Workspace`. You will be asked to pick a channel for the webhook here.
5. Done! A webhook URL will be created.

### How To Get Your Group ID
1. Navigate to https://api.slack.com/methods/usergroups.list/test.
2. Select the correct application under "token".
3. Press "Test Method".
4. Find your group below and copy the value for "ID".

### What to do with Slack Webhook
You can implement the Webhook in one of two ways, either as an environment variable, or as a parameter.

1. In the settings page for your project on CircleCI, click `Environment Variables`. From that page you can click the `Add Variable` button. Finally, enter your webhook as the value, and `SLACK_WEBHOOK` as the name.
2. You can enter the Webhook for the individual status or alert by entering it at the `webhook` parameter, as shown above.

## Contributing
We welcome [issues](https://github.com/CircleCI-Public/slack-orb/issues) to and [pull requests](https://github.com/CircleCI-Public/slack-orb/pulls) against this repository! For further questions/comments about this or other orbs, visit [CircleCI's Orbs discussion forum](https://discuss.circleci.com/c/ecosystem/orbs).

## License
This project is licensed under the MIT License - read [LICENSE](LICENSE) file for details.
