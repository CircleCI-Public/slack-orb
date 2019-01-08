# Slack Orb

Easily integrate custom [Slack](https://slack.com/ "Slack") notifications into your [CircleCI](https://circleci.com/ "CircleCI") projects. Create custom alert messages for any job or receive status updates.

Learn more about [Orbs](https://github.com/CircleCI-Public/config-preview-sdk/blob/master/docs/using-orbs.md "orb").

## Usage

Example config:

```yaml
version: 2.1

orbs:
  slack: circleci/slack@1.0.0

jobs:
  build:
    docker:
      - image: <docker image>
    steps:
      - slack/<command>
```

`slack@1.0.0` from the `circleci` namespace is imported into the config.yml as `slack` and can then be referenced as such in any job or workflow.

## Commands

- ### Notify

|  Usage | slack/notify   |
| ------------ | ------------ |
| **Description:**  | Notify a slack channel with a custom message at any point in a job with this custom step. |
|  **Parameters:** | - **webhook:**  Enter either your Webhook value or use the CircleCI UI to add your token under the `SLACK_WEBHOOK` environment variable <br><br> - **message:** Enter your custom message to send to your Slack channel.  <br> <br> - **mentions:** A comma separated list of Slack user IDs, or Group (SubTeam) IDs. example 'USER1,USER2,USER3'. Note, these are Slack User IDs, not usernames. The user ID can be found on the user's profile. Look below for infomration on obtaining Group ID. <br> <br> - **color:** Color can be set for a notification to help differentiate alerts.|

Example:

```yaml
version: 2.1

orbs:
  slack: circleci/slack@1.0.0

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

See Slack's [Basic message formatting](https://api.slack.com/docs/message-formatting) documentation for guidance on formatting notification messages.

- ### Status

|  Usage | slack/status   |
| ------------ | ------------ |
| **Description:** | Send a status alert at the end of a job based on success or failure. This must be the last step in a job. |
|  **Parameters:** | -  **webhook:** Enter either your Webhook value or use the CircleCI UI to add your token under the `SLACK_WEBHOOK` environment variable <br> <br> - **fail_only:** `false` by default. If set to `true, successful jobs will _not_ send alerts <br> <br> - **mentions:**  comma separated list of Slack user IDs, or Group (SubTeam) IDs. example 'USER1,USER2,USER3'. Note, these are Slack User IDs, not usernames. The user ID can be found on the user's profile. Look below for infomration on obtaining Group ID. <br> <br> - **only_for_branch**: Optional: If set, a specific branch for which status updates will be sent. |

Example:

```yaml
version: 2.1

orbs:
  slack: circleci/slack@1.0.0

jobs:
  build:
    docker:
      - image: <docker image>
    steps:
      # With fail_only set to true, no alert will be sent in this example. Change the exit status on the next line to produce an error.
      - run: exit 0

      - slack/status:
          mentions: "USERID1,USERID2" # Optional: Enter the Slack IDs of any user or group (sub_team) to be mentioned
          fail_only: "true" # Optional: if set to "true" then only failure messages will occur.
          webhook: "webhook" # Optional: Enter a specific webhook here or the default will use $SLACK_WEBHOOK
          only_for_branch: "master" # Optional: If set, a specific branch for which status updates will be sent. In this case, only for pushes to master branch.
```

![Status Success Example](/img/statusSuccess.PNG)
![Status Fail Example](/img/statusFail.PNG)

## Dependencies / Requirements

- ### Bash Shell

Due to the limitations of the `sh` shell, `Bash` is required. `Bash` is the default shell used on CircleCI and the Orb will be compatible with most images. Images such as `Alpine` that do not contain the `Bash` shell by default are incompatible and en error message will be logged. You may install the Bash shell through your package manager (example: `apk add bash`) in the Dockerfile for the image you are using.

- ### cURL

cURL is used to post the Webhook data and must be installed in the container to function properly.

## Help

**How to get your Slack Webhook:**  Full instructions can be found at Slack: https://api.slack.com/incoming-webhooks

1. [Create Slack App](https://api.slack.com/docs/slack-button#register_your_slack_app). This will also be the name of the "user" that posts alerts to Slack. You'll be asked for which Workspace this app belongs to.
2. In the settings for the app, enable `Incoming Webhooks`
3. In the left hand panel of your Slack app settings, under `Features` click `Incoming Webhooks`
4. Click `Add New Webhook to Workspace`. You will be asked to pick a channel for the webhook here.
5. Done! A webhook URL will be created.

**How To Get Your Group ID:**

1. Navigate to https://api.slack.com/methods/usergroups.list/test
2. Select the correct application under "token"
3. Press "Test Method"
4. Find your group below and copy the value for "ID"

**What to do with Slack Webhook:** You can implement the Webhook in one of two ways, as an environment variable, or as a parameter.

1. In the settings page for your project on CircleCI, click `Environment Variables`. From that page you can click the `Add Variable` button. Finally, enter your webhook as the value, and `SLACK_WEBHOOK` as the name.
2. You can enter the Webhook for the individual status or alert by entering is at the `webhook` parameter, as shown above.
