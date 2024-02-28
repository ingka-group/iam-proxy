# IAM Proxy service

## Description

This service defines a simple IAM verification procedure for OAuth2 authentication cycles. It will generate a token for
a valid user with expiration time. It will also provide functionality for the validation of tokens.


## Get Started

This documentation assumes that you have a Go environment already present in your local machine.

### Prerequisites

The service only requires 2 inputs for it to function properly. One being the user access credentials, and the other a
secret token for the token generation process. Both of these can be provided though ENV variables.

```shell
IAM_USERS  = base64.rawEncode(`{"<client_id>": { "client_secret": "<client_secret>", "app_name": "<app_name>" }}`)
IAM_SECRET = demo
```

One can verify the functionality of the service by...
- Making a `POST` request to the `/iam/v1/oauth2/token` endpoint to request a token
- Using this token to make a `POST` request to the `/iam/v1/oauth2/validate` endpoint

NOTE: This needs to be done within the token expiration interval

Check also the [postman collection](/docs/IAM.postman_collection.json) for examples and details.

### Running

You can either run the service directly through your IDE for example, or alternatively, you can use the `Dockerfile`
to build a container image and run it.

```bash
$ make build-image
$ docker run iam-proxy -e IAM_SECRET='demo' -e IAM_USERS='eyIxIjogeyAiY2xpZW50X3NlY3JldCI6ICIyIiwgImFwcF9uYW1lIjogIjMiIH19'
```

### Testing

To run the full set of tests you can execute the following command.

```bash
$ make test
```

### Pre-commit

This project uses pre-commit(https://pre-commit.com/) to integrate code checks used to gate commits.

**NOTE**: When pushing to GitHub, an action runs these same checks. Using `pre-commit` locally ensures these
checks will not fail.

```bash
# required only once
$ pre-commit install
pre-commit installed at .git/hooks/pre-commit

# run checks on all files
$ make pre-commit
```

### Other management tasks

For more information on the available targets, you can access the inline help of the Makefile.

```bash
$ make help
```

or, equivalently,

```bash
$ make
```


## Contributing
Please read [CONTRIBUTING](./CONTRIBUTING.md) for more details about making a contribution to this open source project and ensure that you follow our [CODE_OF_CONDUCT](./CODE_OF_CONDUCT.md).


## Contact
If you have any other issues or questions regarding this project, feel free to contact one of the [code owners/maintainers](.github/CODEOWNERS) for a more in-depth discussion.


## Licence
This open source project is licensed under the "Apache-2.0", read the [LICENCE](./LICENCE.md) terms for more details.
