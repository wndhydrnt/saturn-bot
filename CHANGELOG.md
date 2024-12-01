# Changelog

## [0.16.0](https://github.com/wndhydrnt/saturn-bot/compare/v0.15.1...v0.16.0) (2024-12-01)


### Features

* **server:** Add endpoint to process webhooks sent by GitHub ([#105](https://github.com/wndhydrnt/saturn-bot/issues/105)) ([fa81814](https://github.com/wndhydrnt/saturn-bot/commit/fa818145097314b382521ae4f75a2191404b4c44))
* **server:** Add endpoint to process webhooks sent by GitLab ([#112](https://github.com/wndhydrnt/saturn-bot/issues/112)) ([2cb93b6](https://github.com/wndhydrnt/saturn-bot/commit/2cb93b64ec11af8f97cbb300bee7fb3724beb0a4))
* **server:** Add total number of items to pagination responses ([6a4f722](https://github.com/wndhydrnt/saturn-bot/commit/6a4f722cbab56d86fb8c719216642ac81f96779f))
* **server:** Delay execution of a task when a webhook is received ([#115](https://github.com/wndhydrnt/saturn-bot/issues/115)) ([047169b](https://github.com/wndhydrnt/saturn-bot/commit/047169b5955eca072ba060e10170027da2df7e82))
* **server:** Inspect upcoming and past runs via an endpoint ([#114](https://github.com/wndhydrnt/saturn-bot/issues/114)) ([8ed4ded](https://github.com/wndhydrnt/saturn-bot/commit/8ed4ded3488e444544f4636fada3e6521b7958d5))
* **task:** Inputs ([#107](https://github.com/wndhydrnt/saturn-bot/issues/107)) ([0bffdd4](https://github.com/wndhydrnt/saturn-bot/commit/0bffdd410b049cbc8a212cd7d919f6f8e2c4f140))


### Bug Fixes

* **host:** Auto-merge GitHub PR fails if repository allows only "squash" merge method ([#110](https://github.com/wndhydrnt/saturn-bot/issues/110)) ([57c3673](https://github.com/wndhydrnt/saturn-bot/commit/57c3673a0c40d377c7b1911c5ed167d5a6fa35ff))
* **host:** Too eager to delete GitHub branch ([#111](https://github.com/wndhydrnt/saturn-bot/issues/111)) ([ccb8a6b](https://github.com/wndhydrnt/saturn-bot/commit/ccb8a6b71df7363cfe5aa3fca8e2157f0e1ec1d3))
* **processor:** Pull request not recreated if closed ([#117](https://github.com/wndhydrnt/saturn-bot/issues/117)) ([c554240](https://github.com/wndhydrnt/saturn-bot/commit/c55424090d60ea0fa64ff51f0c6526194814742d))
* **server:** Runs skipped when more than one Task is registered ([#116](https://github.com/wndhydrnt/saturn-bot/issues/116)) ([e3f7956](https://github.com/wndhydrnt/saturn-bot/commit/e3f79560c9fac2732d0f465322058815cddffd27))
* **server:** Wrong run reason "manual" when task changed on disk ([c825e95](https://github.com/wndhydrnt/saturn-bot/commit/c825e95efdfab1e3a9dab45c405148cec1a45b66))
* **template:** Description of pull request states that PR gets merged in 0s ([#108](https://github.com/wndhydrnt/saturn-bot/issues/108)) ([b463fba](https://github.com/wndhydrnt/saturn-bot/commit/b463fba8a97cb21579fd855e7305adecd6772ace))

## [0.15.1](https://github.com/wndhydrnt/saturn-bot/compare/v0.15.0...v0.15.1) (2024-11-11)


### Bug Fixes

* **host:** Do not time out if rate-limiting of GitLab API is active ([#103](https://github.com/wndhydrnt/saturn-bot/issues/103)) ([d475e03](https://github.com/wndhydrnt/saturn-bot/commit/d475e031b9b82d5d91ffda1a0e19794c5ebe782a))

## [0.15.0](https://github.com/wndhydrnt/saturn-bot/compare/v0.14.0...v0.15.0) (2024-11-03)


### ⚠ BREAKING CHANGES

* **filter:** Remove filter jsonpath ([#102](https://github.com/wndhydrnt/saturn-bot/issues/102))

### Features

* **command:** Run command sends metrics to Prometheus Pushgateway ([#98](https://github.com/wndhydrnt/saturn-bot/issues/98)) ([da7df05](https://github.com/wndhydrnt/saturn-bot/commit/da7df05abd831e0219b428dc0c67ca94858ff81d))
* **filter:** Add jq filter ([#101](https://github.com/wndhydrnt/saturn-bot/issues/101)) ([7f6848c](https://github.com/wndhydrnt/saturn-bot/commit/7f6848ce99a1e8b5e8cc7081962dee3c6403a190))
* **filter:** Remove filter jsonpath ([#102](https://github.com/wndhydrnt/saturn-bot/issues/102)) ([6bc7089](https://github.com/wndhydrnt/saturn-bot/commit/6bc7089fa5d59b088a155f1eeb29381c23b4a618))
* **git:** Add metrics to track number and duration of git commands ([#99](https://github.com/wndhydrnt/saturn-bot/issues/99)) ([7e7c7c0](https://github.com/wndhydrnt/saturn-bot/commit/7e7c7c022b8c7e5be7e6a2aeb466c0b83f2104c9))
* **host:** Support "Squash commits" and "Delete source branch" settings of GitLab project ([#97](https://github.com/wndhydrnt/saturn-bot/issues/97)) ([5ce431a](https://github.com/wndhydrnt/saturn-bot/commit/5ce431a67b1a04f06b7ccc94001fe9a14a8652cf))
* Set up shell completion ([4634532](https://github.com/wndhydrnt/saturn-bot/commit/46345322d8dc9387b572bd395333e077e040fa53))
* **task:** Ensure that branch name does not exceed length ([#100](https://github.com/wndhydrnt/saturn-bot/issues/100)) ([bac731d](https://github.com/wndhydrnt/saturn-bot/commit/bac731df77d65bb4485b57d4620b9f98cac56114))

## [0.14.0](https://github.com/wndhydrnt/saturn-bot/compare/v0.13.0...v0.14.0) (2024-10-20)


### Features

* **action:** Pass env var TASK_DIR to script action ([#91](https://github.com/wndhydrnt/saturn-bot/issues/91)) ([955cf5a](https://github.com/wndhydrnt/saturn-bot/commit/955cf5a5c71fe318a6b51708a0bcf8ad0bf1467c))
* **command:** New command "ci" ([#94](https://github.com/wndhydrnt/saturn-bot/issues/94)) ([de968b7](https://github.com/wndhydrnt/saturn-bot/commit/de968b70fee9176757ee911f47e703a435a8eac5))
* **filter:** Specify multiple expressions in filter jsonpath ([#90](https://github.com/wndhydrnt/saturn-bot/issues/90)) ([31d8ebf](https://github.com/wndhydrnt/saturn-bot/commit/31d8ebf018ced15fd947e1ac8b74aa82fdec0969))
* **filter:** Specify multiple expressions in filter xpath ([#89](https://github.com/wndhydrnt/saturn-bot/issues/89)) ([4a96272](https://github.com/wndhydrnt/saturn-bot/commit/4a96272caa57a9e4ca0bf9f95ecacf3ba9cdac5f))
* **host:** Cache files downloaded from a host ([#87](https://github.com/wndhydrnt/saturn-bot/issues/87)) ([263c13b](https://github.com/wndhydrnt/saturn-bot/commit/263c13b6db515ff491211a9f087acff5abc3f3ec))
* **task:** Execute repository filter first ([#92](https://github.com/wndhydrnt/saturn-bot/issues/92)) ([b38c1e9](https://github.com/wndhydrnt/saturn-bot/commit/b38c1e927d5c2cd7794261a256fc1f9e34be20ce))


### Bug Fixes

* **host:** Do not update assignees or reviewers when task does not define them ([#93](https://github.com/wndhydrnt/saturn-bot/issues/93)) ([33a514a](https://github.com/wndhydrnt/saturn-bot/commit/33a514a12cb0afcf9b7e1c189cc4a1f0f23e5c2b))

## [0.13.0](https://github.com/wndhydrnt/saturn-bot/compare/v0.12.0...v0.13.0) (2024-10-13)


### Features

* **docker:** Install JRE in full version of Docker image ([4eb786f](https://github.com/wndhydrnt/saturn-bot/commit/4eb786f490d300e6374f8a755995464706327bdc))
* **filter:** Add jsonpath filter ([#84](https://github.com/wndhydrnt/saturn-bot/issues/84)) ([2123cfe](https://github.com/wndhydrnt/saturn-bot/commit/2123cfe969da5abdace8896a3b068c57d832f170))
* **filter:** Add xpath filter ([#82](https://github.com/wndhydrnt/saturn-bot/issues/82)) ([925421b](https://github.com/wndhydrnt/saturn-bot/commit/925421b72960d9ea1353cbb39c3f7a21025ab322))
* **filter:** Make filter fileContent match against whole content of file ([756e86c](https://github.com/wndhydrnt/saturn-bot/commit/756e86c10d3c40450a6a57a0c433512ab69e873c))


### Bug Fixes

* **config:** Fail early if no GitHub token or GitLab token has been defined ([#86](https://github.com/wndhydrnt/saturn-bot/issues/86)) ([0f6423a](https://github.com/wndhydrnt/saturn-bot/commit/0f6423aec65e870204dab300dff3e078f9dc3a9c))

## [0.12.0](https://github.com/wndhydrnt/saturn-bot/compare/v0.11.1...v0.12.0) (2024-09-28)


### Features

* **action:** Add new action `script` ([#76](https://github.com/wndhydrnt/saturn-bot/issues/76)) ([9e70ed8](https://github.com/wndhydrnt/saturn-bot/commit/9e70ed8a3e43ea2586dc8222a9fef41ff33a6a54))
* **command:** Add commands to test and debug a plugin ([#78](https://github.com/wndhydrnt/saturn-bot/issues/78)) ([7c34635](https://github.com/wndhydrnt/saturn-bot/commit/7c346358fb2caec968f6ff4a85a38621c75b7475))
* **plugin:** Write messages of plugin to log ([#79](https://github.com/wndhydrnt/saturn-bot/issues/79)) ([8c8e670](https://github.com/wndhydrnt/saturn-bot/commit/8c8e6707f36264e14434e1abf3e60106759dabf2))
* **task:** Add schedule setting ([#72](https://github.com/wndhydrnt/saturn-bot/issues/72)) ([3945e1a](https://github.com/wndhydrnt/saturn-bot/commit/3945e1ae68b8242830e4d7eb921e06da2f14325f))
* **task:** Support template in branch name and PR title ([#74](https://github.com/wndhydrnt/saturn-bot/issues/74)) ([4e33cc9](https://github.com/wndhydrnt/saturn-bot/commit/4e33cc94c2f47274d67feaa164d50b810452ff58))


### Bug Fixes

* **log:** Work around warning of zap when `logLevel` is `error` ([#81](https://github.com/wndhydrnt/saturn-bot/issues/81)) ([203815a](https://github.com/wndhydrnt/saturn-bot/commit/203815ae35c1ceded55995209cbdb5f7e5927d6e))
* **task:** Log messages of plugins not formatted correctly ([9a169ed](https://github.com/wndhydrnt/saturn-bot/commit/9a169ed65adc4f9bfaf93eaadae0ac88a5444141))

## [0.11.1](https://github.com/wndhydrnt/saturn-bot/compare/v0.11.0...v0.11.1) (2024-09-01)


### Bug Fixes

* **run:** Do not error when an empty repository has been cloned ([#68](https://github.com/wndhydrnt/saturn-bot/issues/68)) ([0af890f](https://github.com/wndhydrnt/saturn-bot/commit/0af890fd04785a3490e231216a01f5f2fb76c706))
* **server:** Support globbing in paths to task files ([#70](https://github.com/wndhydrnt/saturn-bot/issues/70)) ([7155375](https://github.com/wndhydrnt/saturn-bot/commit/715537518d3e7208ac3b0a7de09401416ced111f))

## [0.11.0](https://github.com/wndhydrnt/saturn-bot/compare/v0.10.0...v0.11.0) (2024-08-28)


### Features

* **task:** Auto-close a pull request after a duration has passed ([#64](https://github.com/wndhydrnt/saturn-bot/issues/64)) ([c9e3c60](https://github.com/wndhydrnt/saturn-bot/commit/c9e3c60a7d58b8f8cda5cfe16cbc88f15435bcc0))
* **task:** Log stdout and stderr of plugin ([#66](https://github.com/wndhydrnt/saturn-bot/issues/66)) ([305d07c](https://github.com/wndhydrnt/saturn-bot/commit/305d07c029ba4fd7f2463a2f2c02931e2e273dda))


### Bug Fixes

* **command:** No information about which task failed ([#67](https://github.com/wndhydrnt/saturn-bot/issues/67)) ([8a5a2aa](https://github.com/wndhydrnt/saturn-bot/commit/8a5a2aae8af1eff534c85046eb08b3f8fd4236b6))

## [0.10.0](https://github.com/wndhydrnt/saturn-bot/compare/v0.9.0...v0.10.0) (2024-08-22)


### Features

* **command:** Align args and flags of command "try" with "run" ([#63](https://github.com/wndhydrnt/saturn-bot/issues/63)) ([f36cdcc](https://github.com/wndhydrnt/saturn-bot/commit/f36cdcc9b3f8d5695609833bef335118788b285f))
* **git:** Discover git author from host ([#55](https://github.com/wndhydrnt/saturn-bot/issues/55)) ([b0d098c](https://github.com/wndhydrnt/saturn-bot/commit/b0d098c1d6e3cd55926d4a3c87b66b42d887e648))


### Bug Fixes

* **action:** line* actions fail if temp directory and data directory aren't on the same device ([#62](https://github.com/wndhydrnt/saturn-bot/issues/62)) ([77d1c2e](https://github.com/wndhydrnt/saturn-bot/commit/77d1c2e620da53caed6bd0f51ee576cd83bf4af1))
* **command:** Get data directory via options in run command to fix nil-pointer ([#59](https://github.com/wndhydrnt/saturn-bot/issues/59)) ([b8e598d](https://github.com/wndhydrnt/saturn-bot/commit/b8e598d4278ad52781459b0d9f9b60f708438c73))
* **log:** Log message and error concatenated at the end of a run ([#61](https://github.com/wndhydrnt/saturn-bot/issues/61)) ([26a70c2](https://github.com/wndhydrnt/saturn-bot/commit/26a70c23f6dbcc5608337f71dbefa3afab80d454))

## [0.9.0](https://github.com/wndhydrnt/saturn-bot/compare/v0.8.1...v0.9.0) (2024-08-21)


### Features

* **command:** Use default data directroy in try command ([#58](https://github.com/wndhydrnt/saturn-bot/issues/58)) ([bf9d5bd](https://github.com/wndhydrnt/saturn-bot/commit/bf9d5bdd50b166dc9b75ac892bae4fc669abba16))
* **git:** Add configuration option to clone via SSH ([#51](https://github.com/wndhydrnt/saturn-bot/issues/51)) ([a11e391](https://github.com/wndhydrnt/saturn-bot/commit/a11e391bc723aaff911ec972e6a61a988ab6ecdc))
* **server:** Configure log level of database library ([069f003](https://github.com/wndhydrnt/saturn-bot/commit/069f0036af00d606286c6eb35993e04d22b82200))

## [0.8.1](https://github.com/wndhydrnt/saturn-bot/compare/v0.8.0...v0.8.1) (2024-08-11)


### Bug Fixes

* **plugin:** Cannot load plugin due to failure during checksum caculation ([#47](https://github.com/wndhydrnt/saturn-bot/issues/47)) ([1070e76](https://github.com/wndhydrnt/saturn-bot/commit/1070e760b96507b260d9e1cb84b9c63fcc60c883))
* **task:** Cannot load tasks from glob paths if no shell is available ([#49](https://github.com/wndhydrnt/saturn-bot/issues/49)) ([e8d07eb](https://github.com/wndhydrnt/saturn-bot/commit/e8d07eb5d1171b14b1106b79c29184849f7534b4))

## [0.8.0](https://github.com/wndhydrnt/saturn-bot/compare/v0.7.0...v0.8.0) (2024-08-08)


### Features

* Add server and worker implementations and commands ([#43](https://github.com/wndhydrnt/saturn-bot/issues/43)) ([f318f78](https://github.com/wndhydrnt/saturn-bot/commit/f318f78e0dae387e08a0b552c7b81209202bef86))
* **plugin:** Upgrade to plugin protocol v0.10.0 ([#44](https://github.com/wndhydrnt/saturn-bot/issues/44)) ([1462b43](https://github.com/wndhydrnt/saturn-bot/commit/1462b436f986d07e2e3b582a7b0d11e11b5309e6))


### Bug Fixes

* **cmd:** Globbing of files passed to `run` command ([#46](https://github.com/wndhydrnt/saturn-bot/issues/46)) ([e977900](https://github.com/wndhydrnt/saturn-bot/commit/e977900f1b6cd99e5b37afd59524f95086d79c42))

## [0.7.0](https://github.com/wndhydrnt/saturn-bot/compare/v0.6.0...v0.7.0) (2024-07-28)


### Features

* Add `or` operator to `file` filter ([#35](https://github.com/wndhydrnt/saturn-bot/issues/35)) ([77205a4](https://github.com/wndhydrnt/saturn-bot/commit/77205a4962304f5f1eb272c9de6caf184cf3865a))
* **config:** Allow configuration of paths to Java and Python executables ([#38](https://github.com/wndhydrnt/saturn-bot/issues/38)) ([a8ef3c7](https://github.com/wndhydrnt/saturn-bot/commit/a8ef3c79f2d3e99c1e4d6c203bee1dda20a30266))
* **docker:** Set up a Python virtual environment ([#37](https://github.com/wndhydrnt/saturn-bot/issues/37)) ([1152d50](https://github.com/wndhydrnt/saturn-bot/commit/1152d5063d6419b67cca3707c18e7ce9c5a377c9))
* **task:** Add toggle to activate/deactivate a task ([#39](https://github.com/wndhydrnt/saturn-bot/issues/39)) ([9907dac](https://github.com/wndhydrnt/saturn-bot/commit/9907dac769cec5b430b5f5cfb5b87c96c6914631))


### Bug Fixes

* **task:** Relative path to plugin not resolved ([#40](https://github.com/wndhydrnt/saturn-bot/issues/40)) ([7147e2a](https://github.com/wndhydrnt/saturn-bot/commit/7147e2a82406ff42d51422202cb2eefcd6b1d78e))

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
