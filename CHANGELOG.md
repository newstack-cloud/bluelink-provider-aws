# Changelog

## 0.1.0 (2025-07-22)


### Features

* add alias data source implementation ([365b8b7](https://github.com/newstack-cloud/bluelink-provider-aws/commit/365b8b7a90c2f6f3e7c0c4781132958dc89284dd))
* add alias resource implementation ([d8c4188](https://github.com/newstack-cloud/bluelink-provider-aws/commit/d8c418825805f778391df6a038e950c02b9baaf4))
* add code signing config data source implementation ([97944c2](https://github.com/newstack-cloud/bluelink-provider-aws/commit/97944c2c8f5b660cf45b34280c8c3949f33eb094))
* add code signing config resource implementation ([60a2f02](https://github.com/newstack-cloud/bluelink-provider-aws/commit/60a2f02cad1acece43dcbcd4ed93e692f8a8639d))
* add event source mapping resource implementation ([903c912](https://github.com/newstack-cloud/bluelink-provider-aws/commit/903c91292f821ca31215f3b57088fbafb44d6cf2))
* add function url resource implementation ([2755b27](https://github.com/newstack-cloud/bluelink-provider-aws/commit/2755b27ba8abc7612d4a0bbdf1c36024f9355c08))
* add function version resource implementation ([6a531a7](https://github.com/newstack-cloud/bluelink-provider-aws/commit/6a531a725e96817ff043fb4f62a184db8e7d5a98))
* add iam access key resource implementation ([151dc7d](https://github.com/newstack-cloud/bluelink-provider-aws/commit/151dc7d58173a136691e21b2435700ec1272b3ef))
* add iam group resource implementation ([a36e60b](https://github.com/newstack-cloud/bluelink-provider-aws/commit/a36e60b60157901ad31f2439ff89c9b4b57f1b49))
* add iam instance profile resource implementation ([b3158a1](https://github.com/newstack-cloud/bluelink-provider-aws/commit/b3158a178a1104c00fce9372faa1eb14c2c77d12))
* add iam managed policy resource ([f76fedc](https://github.com/newstack-cloud/bluelink-provider-aws/commit/f76fedc1aa9b43967c1bdca45dfb6beb370c620f))
* add iam oidc provider resource ([dd4a8b9](https://github.com/newstack-cloud/bluelink-provider-aws/commit/dd4a8b9db1b437b7eba1203b81e4b838dd5042bf))
* add iam role resource implementation ([3e0fb0e](https://github.com/newstack-cloud/bluelink-provider-aws/commit/3e0fb0ecd39706404d4ac9558f4624873289e764))
* add iam saml provider resource implementation ([a9936cf](https://github.com/newstack-cloud/bluelink-provider-aws/commit/a9936cf4b3bd331f289daa1bbc60c98a484a1526))
* add iam server certificate resource ([51f95a8](https://github.com/newstack-cloud/bluelink-provider-aws/commit/51f95a861a3515b67ed0ee3af688062f3ae69f74))
* add implementation of lambda function version resource ([11c8efb](https://github.com/newstack-cloud/bluelink-provider-aws/commit/11c8efb8fe7c28f132d594df98613613cf27dc46))
* add implementation of lambda layer version resource ([19e6a6a](https://github.com/newstack-cloud/bluelink-provider-aws/commit/19e6a6a56ebb10887f601abae63debb1253472b6))
* add implementation of the lambda function resource ([7a519d9](https://github.com/newstack-cloud/bluelink-provider-aws/commit/7a519d9d5a67d07156a7e496ab3b2c5e1bca55ad))
* add lambda alias resource implementation ([21d825f](https://github.com/newstack-cloud/bluelink-provider-aws/commit/21d825f75d8233781067824989c54045db0946d4))
* add lambda code signing config resource implementation ([10ae1ca](https://github.com/newstack-cloud/bluelink-provider-aws/commit/10ae1ca04a41c32ebb821557eb6919c0a7074c89))
* add lambda event invoke config resource implementation ([c00add2](https://github.com/newstack-cloud/bluelink-provider-aws/commit/c00add26156940f5e2f8ba86ee84febbbe99ea0e))
* add lambda function to code signing config link implementation ([1990c78](https://github.com/newstack-cloud/bluelink-provider-aws/commit/1990c7885ab8562bc25c7d24ab876e5cac0e1dda))
* add lambda function url data source implementation ([9a34b1e](https://github.com/newstack-cloud/bluelink-provider-aws/commit/9a34b1e7c58186aad5fc234f114b15f7381ebe25))
* add lambda layer version data source implementation ([b49be69](https://github.com/newstack-cloud/bluelink-provider-aws/commit/b49be69f833969592bdc7ccd47fc37990c6efa8c))
* add lambda layer version permission resource implementation ([6dede22](https://github.com/newstack-cloud/bluelink-provider-aws/commit/6dede228de7160098649914dc70eb425f8cc95f4))
* add missing perm boundary updates and tag sorting ([89e7ac6](https://github.com/newstack-cloud/bluelink-provider-aws/commit/89e7ac682bb3c08e2496dd903d84b5740e50ca0b))
* add missing tag and policy diff behaviour for roles ([b3260a6](https://github.com/newstack-cloud/bluelink-provider-aws/commit/b3260a6af5c16a297641910832ea3d262f152e53))
* update function to csc link to return resource data mappings ([073dead](https://github.com/newstack-cloud/bluelink-provider-aws/commit/073dead45a7b9339d83f1805808acbb02142fb87))


### Bug Fixes

* add corrections to function version examples ([690f31a](https://github.com/newstack-cloud/bluelink-provider-aws/commit/690f31a44576c97a959da59c6c3ee517166c28b4))
* add missing computed fields from oidc provider update response ([f542b5d](https://github.com/newstack-cloud/bluelink-provider-aws/commit/f542b5d2c0640e769cc509d3adedcb08617a9cca))
* add missing nil checks and a full test suite for get external state ([4c09ef0](https://github.com/newstack-cloud/bluelink-provider-aws/commit/4c09ef028f239237cde84c9d347b870484ab02e1))
* correct examples for lambda layer version resource ([dc57ae9](https://github.com/newstack-cloud/bluelink-provider-aws/commit/dc57ae94ebb6bdd793ff20ee475e49348082a100))
* correct formatting in layer version get external state file ([9bd563e](https://github.com/newstack-cloud/bluelink-provider-aws/commit/9bd563e1f0186255522aead60d723f6d6020bcc7))
* ensure iam resources are registered with the plugin provider ([505f2b6](https://github.com/newstack-cloud/bluelink-provider-aws/commit/505f2b68b49025e941037f6a3cd78feee864398a))
* remove zipFile field from lambda layer version ([6f1aba5](https://github.com/newstack-cloud/bluelink-provider-aws/commit/6f1aba58cbd08296faca83b1aa4d03e07eba3879))

## Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
