# Contributing Guide

We love your input! We want to make contributing to this project as easy and transparent as possible, whether it's:

- Reporting a bug
- Discussing the current state of the code
- Submitting a fix
- Proposing new features
- Becoming a maintainer

## We Develop with Github
We use github to host code, to track issues and feature requests, as well as accept pull requests.

## We Use [Github Flow](https://guides.github.com/introduction/flow/index.html), So All Code Changes Happen Through Pull Requests
Pull requests are the best way to propose changes to the codebase. We actively welcome your pull requests:

1. Fork the repo and create your branch from `main`.
2. If you've added code that should be tested, add tests.
3. If you've changed APIs, update the documentation.
4. Ensure the test suite passes.
5. Make sure your code lints.
6. Issue that pull request!

## Any contributions you make will be under the MIT Software License
In short, when you submit code changes, your submissions are understood to be under the same [MIT License](http://choosealicense.com/licenses/mit/) that covers the project. Feel free to contact the maintainers if that's a concern.

## Report bugs using Github's [issues](https://github.com/briandk/transcriptase-atom/issues)
We use GitHub issues to track public bugs. Report a bug by [opening a new issue](https://github.com/briandk/transcriptase-atom/issues/new); it's that easy!

## Write bug reports with detail, background, and sample code
Great Bug Reports tend to have:

- A quick summary and/or background
- Steps to reproduce
- Be specific!
- Give sample code if you can.
- What you expected would happen
- What actually happens
- Notes (possibly including why you think this might be happening, or stuff you tried that didn't work)

## Use a Consistent Coding Style
We're flexible about coding style, but we do ask that you remain consistent with the styles throughout the code. 

## License
By contributing, you agree that your contributions will be licensed under its MIT License.

## References
This document was adapted from the open-source contribution guidelines for [Facebook's Draft](https://github.com/facebook/draft-js)

## Debugging Locally

The [launch.json](.vscode/launch.json) file has been configured to debug the application locally. The values for the required global arguments are passed via the .env.sh file. The following is a sample .env.sh file:

```shell
MPF_SUBSCRIPTIONID=YOUR_SUBSCRIPTION_ID
MPF_TENANTID=YOUR_TENANT_ID
MPF_SPCLIENTID=YOUR_SP_CLIENT_ID
MPF_SPCLIENTSECRET=YOUR_SP_CLIENT_SECRET
MPF_SPOBJECTID=YOUR_SP_OBJECT_ID
```

Next the arguments can me modified in the launch.json file for ARM or Terraform as shown:

* ARM

```json
"args": ["arm", "--templateFilePath", "PATH_TO_ARM_TEMPLATE_FILE", "--parametersFilePath", "PATH_TO_PARAMETERS_FILE"]
```

* Terraform

```json
"args": ["terraform", "--tfPath", "PATH_TO_TERRAFORM_EXECUTABLE" , "--workingDir", "PATH_TO_TERRAFORM_MODULE_DIRECTORY", "--varFilePath", "PATH_TO_TERRAFORM_VARIABLES_FILE"]
```

These values need to be modified to match your environment. You can also modify the launch.json to change the logging verbosity, which is debug by default while debugging.

