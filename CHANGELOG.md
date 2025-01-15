# Changelog

## [0.1.0](https://github.com/statnett/provider-cloudian/compare/v0.0.6...v0.1.0) (2025-01-15)


### ⚠ BREAKING CHANGES

* 1:1 between MR and infra for ID fields ([#115](https://github.com/statnett/provider-cloudian/issues/115))
* ref user in access key ([#112](https://github.com/statnett/provider-cloudian/issues/112))
* use Crossplane reference to connect User to Group ([#104](https://github.com/statnett/provider-cloudian/issues/104))
* set Cloudian Group and User ids from Crossplane external-name ([#102](https://github.com/statnett/provider-cloudian/issues/102))

### Features

* use Crossplane reference to connect User to Group ([#104](https://github.com/statnett/provider-cloudian/issues/104)) ([9a17ac1](https://github.com/statnett/provider-cloudian/commit/9a17ac1f4b708d3385f5e4572a2dee85c61bd745))


### Bug Fixes

* **deps:** update module sigs.k8s.io/controller-tools to v0.17.1 ([#99](https://github.com/statnett/provider-cloudian/issues/99)) ([84f2c6e](https://github.com/statnett/provider-cloudian/commit/84f2c6e2b713d9e4211f89609372e8118b0d6e07))
* **sdk:** create creds ([#114](https://github.com/statnett/provider-cloudian/issues/114)) ([84f5f25](https://github.com/statnett/provider-cloudian/commit/84f5f2502491ed8e0e078cd404007af6f54f055c))
* **sdk:** error handling ([#113](https://github.com/statnett/provider-cloudian/issues/113)) ([9aba614](https://github.com/statnett/provider-cloudian/commit/9aba614cd7f14b9d417a9a648b2b11be3c2b50cf))


### Miscellaneous Chores

* 1:1 between MR and infra for ID fields ([#115](https://github.com/statnett/provider-cloudian/issues/115)) ([445e881](https://github.com/statnett/provider-cloudian/commit/445e8819ed68323f426e58a5174a9251abd07a44))
* ref user in access key ([#112](https://github.com/statnett/provider-cloudian/issues/112)) ([02617c2](https://github.com/statnett/provider-cloudian/commit/02617c2aa6f7f72ecf5c3232f366b13deeff7b19))
* set Cloudian Group and User ids from Crossplane external-name ([#102](https://github.com/statnett/provider-cloudian/issues/102)) ([5c60527](https://github.com/statnett/provider-cloudian/commit/5c605277674ad17607fb05f5934c5131960de3f5))

## [0.0.6](https://github.com/statnett/provider-cloudian/compare/v0.0.5...v0.0.6) (2025-01-14)


### Features

* connection config.toml in connection details ([#100](https://github.com/statnett/provider-cloudian/issues/100)) ([6d0735c](https://github.com/statnett/provider-cloudian/commit/6d0735c753e5e002f9c913c6d29565bd42a37b6c))
* reconcile AccessKey ([#93](https://github.com/statnett/provider-cloudian/issues/93)) ([79bf9b9](https://github.com/statnett/provider-cloudian/commit/79bf9b955b8f470bf29b4a25c0847a2f86e8a2b2))

## [0.0.5](https://github.com/statnett/provider-cloudian/compare/v0.0.4...v0.0.5) (2025-01-09)


### Features

* **sdk:** delete user credentials ([#90](https://github.com/statnett/provider-cloudian/issues/90)) ([3a0176d](https://github.com/statnett/provider-cloudian/commit/3a0176d412a4e64156cebd0c991b3f2241ea5b38))
* **sdk:** get use credentials ([#89](https://github.com/statnett/provider-cloudian/issues/89)) ([cdfa16f](https://github.com/statnett/provider-cloudian/commit/cdfa16fa047d53427ef57c8554ba963e9e27356f))


### Bug Fixes

* **sdk:** delete group non-success handling ([#91](https://github.com/statnett/provider-cloudian/issues/91)) ([2228b48](https://github.com/statnett/provider-cloudian/commit/2228b4892eb1644db5425c9bb08097bee453b4b7))

## [0.0.4](https://github.com/statnett/provider-cloudian/compare/v0.0.3...v0.0.4) (2025-01-09)


### Features

* bootstrap empty AccessKey kind ([#86](https://github.com/statnett/provider-cloudian/issues/86)) ([a443195](https://github.com/statnett/provider-cloudian/commit/a443195ceb14f0e2a038bbf73ea772d952f2bc54))
* **sdk:** create credentials ([#85](https://github.com/statnett/provider-cloudian/issues/85)) ([1588392](https://github.com/statnett/provider-cloudian/commit/1588392864b478421be6ee3480cba7b59a5ff918))

## [0.0.3](https://github.com/statnett/provider-cloudian/compare/v0.0.2...v0.0.3) (2025-01-08)


### Bug Fixes

* **deps:** update module sigs.k8s.io/controller-runtime to v0.19.4 ([#80](https://github.com/statnett/provider-cloudian/issues/80)) ([df4e231](https://github.com/statnett/provider-cloudian/commit/df4e2311f62e1a5a5614b122e76a39503a01301e))

## [0.0.2](https://github.com/statnett/provider-cloudian/compare/v0.0.1...v0.0.2) (2025-01-08)


### Features

* fetch user credentials in sdk ([#78](https://github.com/statnett/provider-cloudian/issues/78)) ([ad0ef0a](https://github.com/statnett/provider-cloudian/commit/ad0ef0aae9e0e018d49c9d228fed8a70e68f754e))

## [0.0.1](https://github.com/statnett/provider-cloudian/compare/v0.0.0...v0.0.1) (2025-01-02)


### Features

* add API-resource Group ([#17](https://github.com/statnett/provider-cloudian/issues/17)) ([2321631](https://github.com/statnett/provider-cloudian/commit/232163123cf6493621912a9f5b43ff7dba2d204e))
* add cloudian-sdk with CRUD on group ([#20](https://github.com/statnett/provider-cloudian/issues/20)) ([bf40646](https://github.com/statnett/provider-cloudian/commit/bf40646ad28e6db9139a5b6282164a6ab68704f1))
* add fields to group resource ([#23](https://github.com/statnett/provider-cloudian/issues/23)) ([fea3676](https://github.com/statnett/provider-cloudian/commit/fea36763dcc967019bb7911ceb2ab050bf809cf9))
* add renovate bot config ([#24](https://github.com/statnett/provider-cloudian/issues/24)) ([7a3cba2](https://github.com/statnett/provider-cloudian/commit/7a3cba28d6762916332e76ceaa5ed5c67b7bc4bd))
* allow unspecified group fields ([#47](https://github.com/statnett/provider-cloudian/issues/47)) ([0293409](https://github.com/statnett/provider-cloudian/commit/0293409c25aea9cabe6da0d26b205777e1f01011))
* **cloudian-sdk:** add cloudian user ([#61](https://github.com/statnett/provider-cloudian/issues/61)) ([b003f75](https://github.com/statnett/provider-cloudian/commit/b003f75028a6e494cc08382dd9111e6733eb84af))
* group api defaults ([#49](https://github.com/statnett/provider-cloudian/issues/49)) ([564dd10](https://github.com/statnett/provider-cloudian/commit/564dd10477518690fd32b0914f3921091f575584))
* group controller uses cloudian sdk ([#57](https://github.com/statnett/provider-cloudian/issues/57)) ([13a6161](https://github.com/statnett/provider-cloudian/commit/13a616103ad72518dbe5e3d66f823fb1422e78d5))
* provider-config - connect sdk ([#45](https://github.com/statnett/provider-cloudian/issues/45)) ([2cc661c](https://github.com/statnett/provider-cloudian/commit/2cc661cb175741122c0ee1473673192cce3ac977))
* reconcile user ([#64](https://github.com/statnett/provider-cloudian/issues/64)) ([13f529c](https://github.com/statnett/provider-cloudian/commit/13f529c43d1531260fb1b34c082527c0c7e11d9f))


### Bug Fixes

* add omitempty to optional group fields ([#41](https://github.com/statnett/provider-cloudian/issues/41)) ([12a176b](https://github.com/statnett/provider-cloudian/commit/12a176ba5415250d6a1167b69ddad369ea4e0684))
* bump Go version ([#39](https://github.com/statnett/provider-cloudian/issues/39)) ([1f3f968](https://github.com/statnett/provider-cloudian/commit/1f3f96850438870202ac04d605200d1073d6f600))
* bump outdated Go dependencies ([#69](https://github.com/statnett/provider-cloudian/issues/69)) ([6a93f41](https://github.com/statnett/provider-cloudian/commit/6a93f417f84a9b6d3abd86cf10b0c4337757bff5))
* bump various dependencies ([#31](https://github.com/statnett/provider-cloudian/issues/31)) ([7c7b204](https://github.com/statnett/provider-cloudian/commit/7c7b204e48c8a00610e1ea1365676a604becd231))
* contain entire auth header in ProviderConfig ([#51](https://github.com/statnett/provider-cloudian/issues/51)) ([f8a8b0d](https://github.com/statnett/provider-cloudian/commit/f8a8b0d4550e425f30ef0604fd70ce5de33cc4f4))
* correct cloudian-verb for create and update group ([#44](https://github.com/statnett/provider-cloudian/issues/44)) ([61a2330](https://github.com/statnett/provider-cloudian/commit/61a23308b11d3bfa2a9a500c4d41afac68d709d5))
* **deps:** update module sigs.k8s.io/controller-tools to v0.17.0 ([#70](https://github.com/statnett/provider-cloudian/issues/70)) ([a6a2d91](https://github.com/statnett/provider-cloudian/commit/a6a2d918e1f127fcb5cf7b130a18fc5d148c8f08))
* remove mentions of defaulting ([#59](https://github.com/statnett/provider-cloudian/issues/59)) ([ae2e679](https://github.com/statnett/provider-cloudian/commit/ae2e679716fc0aa5772490eef58a89dd2f0b6344))
* repo location ([35536aa](https://github.com/statnett/provider-cloudian/commit/35536aa67e52ea2644bbce3a304348a08314df25))
* user resource exists when it is upToDate ([#65](https://github.com/statnett/provider-cloudian/issues/65)) ([5af324b](https://github.com/statnett/provider-cloudian/commit/5af324baf3a7851109acccb01e5d4b7fca1c5722))
