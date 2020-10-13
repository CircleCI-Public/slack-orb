# Slack Orb  [![CircleCI Build Status](https://circleci.com/gh/CircleCI-Public/slack-orb.svg?style=shield "CircleCI Build Status")](https://circleci.com/gh/CircleCI-Public/slack-orb) [![CircleCI Orb Version](https://img.shields.io/badge/endpoint.svg?url=https://badges.circleci.io/orb/circleci/slack)](https://circleci.com/orbs/registry/orb/circleci/slack) [![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/circleci-public/slack-orb/master/LICENSE) [![CircleCI Community](https://img.shields.io/badge/community-CircleCI%20Discuss-343434.svg)](https://discuss.circleci.com/c/ecosystem/orbs)

Send Slack notifications from your CircleCI pipelines even easier with Slack Orb 4.0

[What are Orbs?](https://circleci.com/orbs/)

## Usage

### Setup

In order to use the Slack Orb on CircleCI you will need to create a Slack App and provide an OAuth token. Find the guide in the wiki: [How to setup slack orb](https://github.com/CircleCI-Public/slack-orb/wiki/Setup)

### Use In Config

For full usage guidelines, see the [orb registry listing](http://circleci.com/orbs/registry/orb/circleci/slack).

## Templates

The Slack Orb comes with a number of included templates to get your started with minimal setup. Feel free to use an included template or create your own.

| Template Preview  | Template  | Description |
| ------------- | ------------- | ------------- |
| ![basic_fail_1](./.github/img/basic_fail_1.png)  | basic_fail_1   | Should be used with the "fail" event. |
| ![success_tagged_deploy_1](./.github/img/success_tagged_deploy_1.png)  | success_tagged_deploy_1   | To be used in the event of a successful deployment job. _see orb usage examples_ |
| ![basic_on_hold_1](./.github/img/basic_on_hold_1.png)  | basic_on_hold_1   | To be used in the on-hold job. _see orb [usage examples](https://circleci.com/developer/orbs/orb/circleci/slack#usage-examples)_  |


## Custom Message Template

  1. Open the Slack Block Kit Builder: https://app.slack.com/block-kit-builder/
  2. Design your desired notification message.
  3. Replace any placeholder values with $ENV environment variable strings.
  4. Set the resulting code as the value for your `custom` parameter.

  ```yaml
- slack/notify:
      event: always
      custom: |
        {
          "blocks": [
            {
              "type": "section",
              "fields": [
                {
                  "type": "plain_text",
                  "text": "*This is a text notification*",
                  "emoji": true
                }
              ]
            }
          ]
        }
  ```


## Branch Filtering

Limit Slack notifications to particular branches with the "branch_pattern" parameter.

```
A comma separated list of regex matchable branch names. Notifications will only be sent if sent from a job from these branches. By default ".+" will be used to match all branches. Pattern must match the full string, no partial matches.
```

See [usage examples](https://circleci.com/developer/orbs/orb/circleci/slack#usage-examples).

---

## FAQ

View the [FAQ in the wiki](https://github.com/CircleCI-Public/slack-orb/wiki/FAQ)

## Contributing

We welcome [issues](https://github.com/CircleCI-Public/slack-orb/issues) to and [pull requests](https://github.com/CircleCI-Public/slack-orb/pulls) against this repository!

For further questions/comments about this or other orbs, visit [CircleCI's orbs discussion forum](https://discuss.circleci.com/c/orbs).
