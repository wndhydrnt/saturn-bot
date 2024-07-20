# Changelog

## [0.6.0](https://github.com/wndhydrnt/saturn-bot/compare/v0.5.0...v0.6.0) (2024-07-20)


### Features

* Make gitAuthor configuration setting optional ([#29](https://github.com/wndhydrnt/saturn-bot/issues/29)) ([2dc68c7](https://github.com/wndhydrnt/saturn-bot/commit/2dc68c760036109e86d77aa1599497b192162615))


### Bug Fixes

* Remove a chatty log statement ([6ea5cd6](https://github.com/wndhydrnt/saturn-bot/commit/6ea5cd6b8fdbfe4df239a3fdd101c22c36af812f))

## [0.5.0](https://github.com/wndhydrnt/saturn-bot/compare/v0.4.1...v0.5.0) (2024-07-14)


### Features

* Implement change limit feature ([#27](https://github.com/wndhydrnt/saturn-bot/issues/27)) ([dabaef0](https://github.com/wndhydrnt/saturn-bot/commit/dabaef0796257e39f708376db855e187eee025a7))
* Set maximum open pull requests of a Task ([#26](https://github.com/wndhydrnt/saturn-bot/issues/26)) ([ba65adb](https://github.com/wndhydrnt/saturn-bot/commit/ba65adb2531ae8cd8076039b8dc843ef83b99c62))
* Support Java plugins ([#23](https://github.com/wndhydrnt/saturn-bot/issues/23)) ([233b103](https://github.com/wndhydrnt/saturn-bot/commit/233b1034a0f38b14ba08a19539ac4eaec532b73f))


### Bug Fixes

* exec action fails when the name of a command is set ([#25](https://github.com/wndhydrnt/saturn-bot/issues/25)) ([1851b8d](https://github.com/wndhydrnt/saturn-bot/commit/1851b8d79a40ec8ef40c818ffb053e5402a2c78c))

## [0.4.1](https://github.com/wndhydrnt/saturn-bot/compare/v0.4.0...v0.4.1) (2024-07-01)


### Bug Fixes

* Release of full Docker image fails ([4da8e47](https://github.com/wndhydrnt/saturn-bot/commit/4da8e474be431e42be1b58944b0371676ccef768))

## [0.4.0](https://github.com/wndhydrnt/saturn-bot/compare/v0.3.0...v0.4.0) (2024-07-01)


### Features

* Add exec action ([#17](https://github.com/wndhydrnt/saturn-bot/issues/17)) ([5fa1542](https://github.com/wndhydrnt/saturn-bot/commit/5fa154226d2b93653cfc3f71a6f8a9104fc51385))
* Supply repositories to apply Tasks via command-line argument ([#20](https://github.com/wndhydrnt/saturn-bot/issues/20)) ([5d2faa8](https://github.com/wndhydrnt/saturn-bot/commit/5d2faa8e1ebb56b54a1b4e158bbae0ae68e32c60))

## [0.3.0](https://github.com/wndhydrnt/saturn-bot/compare/v0.2.1...v0.3.0) (2024-06-16)


### Features

* Set reviewers of pull request ([#13](https://github.com/wndhydrnt/saturn-bot/issues/13)) ([99c5168](https://github.com/wndhydrnt/saturn-bot/commit/99c51683d50051d881dee1f0dbd26dab8b584759))

## [0.2.1](https://github.com/wndhydrnt/saturn-bot/compare/v0.2.0...v0.2.1) (2024-05-24)


### Bug Fixes

* Unable to set address of GitLab host ([aca60b7](https://github.com/wndhydrnt/saturn-bot/commit/aca60b77ff48a278760815f372fd9f3784fabd6c))

## [0.2.0](https://github.com/wndhydrnt/saturn-bot/compare/v0.1.0...v0.2.0) (2024-05-20)


### Features

* Implement parameter `mode` of action `fileCreate` ([#7](https://github.com/wndhydrnt/saturn-bot/issues/7)) ([612f007](https://github.com/wndhydrnt/saturn-bot/commit/612f007c09cedf3b717714527927cf61a28dcd05))
* Pass data between plugins ([#10](https://github.com/wndhydrnt/saturn-bot/issues/10)) ([406ffd0](https://github.com/wndhydrnt/saturn-bot/commit/406ffd0d5a599df205dbe04d522a01f89ffde7b6))
* Pass template variables from plugins to templates ([#9](https://github.com/wndhydrnt/saturn-bot/issues/9)) ([7559810](https://github.com/wndhydrnt/saturn-bot/commit/755981038a75b0ea1c6bfefa3ed89dc3a0583f2a))
* Send data on pull request to plugin ([#8](https://github.com/wndhydrnt/saturn-bot/issues/8)) ([e29cc08](https://github.com/wndhydrnt/saturn-bot/commit/e29cc08a4f5075e9c18c4b39f94e7db0500f7af1))
* Set assignees on pull requests ([#6](https://github.com/wndhydrnt/saturn-bot/issues/6)) ([87c0e03](https://github.com/wndhydrnt/saturn-bot/commit/87c0e037c8cefaef49505e293c6474001eef3c13))


### Bug Fixes

* Filter "repository" does not match any GitLab repositories ([e9c54ce](https://github.com/wndhydrnt/saturn-bot/commit/e9c54ce14e335feca7d44532d3214a2a19f44bc5))

## 0.1.0 (2024-05-14)


### Features

* Add emojis to PR template ([de3c95c](https://github.com/wndhydrnt/saturn-bot/commit/de3c95c59af8e1e33665ac981e63b9415a3e0416))
* Rename "saturn-sync" to "saturn-bot" ([c3fe343](https://github.com/wndhydrnt/saturn-bot/commit/c3fe343c5311f9a5607c4591b6d4b7a34ae88253))


### Bug Fixes

* Don't update cache when running in dry-run mode ([b693225](https://github.com/wndhydrnt/saturn-bot/commit/b69322595a24255fa29a7c3c80496e9c44b1788a))
* Make action "file" create directory if it does not exist ([63cc495](https://github.com/wndhydrnt/saturn-bot/commit/63cc495fe21afd0eb6031a9dd7dea4ff741f0346))
* Make path to task absolute when processing it ([134e46b](https://github.com/wndhydrnt/saturn-bot/commit/134e46bdcc337bacbef9394fbc464a72cef67514))


### Miscellaneous Chores

* prepare first release ([3509561](https://github.com/wndhydrnt/saturn-bot/commit/35095610a19b94b73d9172445b103fe1ebcd2fef))
