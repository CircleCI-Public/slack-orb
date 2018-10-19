# Slack Orb


Easily integrate custom [Slack](https://slack.com/ "Slack") notifications into your [CircleCI](https://circleci.com/ "CircleCI") projects. Create custom alert messages for any job or receive status updates. 

Learn more about [orbs](https://github.com/CircleCI-Public/config-preview-sdk/blob/master/docs/using-orbs.md "orb").


## Usage

Example config:
```yaml
orbs:
  slack: circleci/slack@1.0.0

jobs:
  build:
    docker: 
      - image: <docker image>
    steps:
      - slack/<command>

```
`slack@dev:<version>` from the `sandbox` namespace is imported into `slack` which can then be referenced in a step in any job you require.

## Commands
- ### Notify

|  Usage | slack/notify   |
| ------------ | ------------ |
| **Description:**  | Notify a slack channel with a custom message  |   
|  **Parameters:** | - **webhook:**  Enter either your Webhook value or use the CircleCI UI to add your token under the `SLACK_WEBHOOK` environment variable <br><br> - **message:** Enter your custom message to send to your Slack channel.  <br> <br> - **mentions:** A comma separated list of Slack user IDs. example 'USER1,USER2,USER3'. Note, these are Slack User IDs, not usernames. The user ID can be found on the user's profile. <br> <br> - **color:** Color can be set for a notification to help differentiate alerts.|

Example:

```yaml
jobs:
  alertme:
    docker: 
      - image: circleci/node
    steps:
      - slack/notify:
        message: "This is a custom message notification" #Enter your own message
        mentions: "USERID1,USERID2" #Enter the Slack IDs of any users who should be alerted to this message.
        color: "#42e2f4" #Assign custom colors for each notification
        webhook: "webhook" #Enter a specific webhook here or the default will use $SLACK_WEBHOOK 
```
![Custom Message Example](/img/notifyMessage.PNG)

- ### Status

|  Usage | slack/status   |
| ------------ | ------------ |
| **Description:**  | Send a status alert at the end of a job based on success or failure. Must be last step in job  |   
|  **Parameters:** | -  **webhook:** Enter either your Webhook value or use the CircleCI UI to add your token under the `SLACK_WEBHOOK` environment variable <br> <br> -  **fail_only:** `false` by default. If set to `true, successful jobs will _not_ send alerts <br> <br> - **mentions:** A comma separated list of Slack user IDs. example 'USER1,USER2,USER3'. Note, these are Slack User IDs, not usernames. The user ID can be found on the user's profile. | 

Example:

```yaml
jobs:
  alertme:
    docker: 
      - image: circleci/node
    steps:
      - slack/notify:
        mentions: "USERID1,USERID2" #Enter the Slack IDs of any users who should be alerted to this message.
        fail_only: "true" #Optional: if set to "true" then only failure messages will occur.
        webhook: "webhook" #Enter a specific webhook here or the default will use $SLACK_WEBHOOK 
```

![Status Success Example](/img/statusSuccess.PNG)
![Status Fail Example](/img/statusFail.PNG)
