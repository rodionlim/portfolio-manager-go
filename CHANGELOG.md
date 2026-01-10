## [1.33.1] - 2026-01-10

### 🐛 Bug Fixes

- *(ui,blotter)* Ensure that mobile view buttons are in right positions

## [1.33.0] - 2026-01-10

### 🚀 Features

- Implement CLI backup and restore functionality with version command

## [1.32.0] - 2026-01-10

### 🚀 Features

- *(ui,blotter)* Allow for new columns on ref data
- Allow proper filters for trade date

## [1.31.2] - 2025-12-27

### 🐛 Bug Fixes

- *(ui)* Update volatility calculation #140

## [1.31.1] - 2025-12-27

### 🐛 Bug Fixes

- *(ui)* Change volatility calculation #140

## [1.31.0] - 2025-12-27

### 🚀 Features

- *(ui)* Add volatility metrics

## [1.30.1] - 2025-12-20

### 🐛 Bug Fixes

- *(llm)* Force analysis did not work
- Import cycle in tests

## [1.30.0] - 2025-12-20

### 🚀 Features

- *(seed)* Add jardine matheson to seed
- *(llm)* Update default model to gemini 3 flash

## [1.29.1] - 2025-12-01

### 🐛 Bug Fixes

- *(ui)* Handle sector navigation to top 10 stocks view additional mapping #134

## [1.29.0] - 2025-11-30

### 🚀 Features

- *(analytics)* Add portfolio filter, sector navigation, and lightweight positions endpoint

## [1.28.0] - 2025-11-27

### 🚀 Features

- *(seed)* Add sgx to seed data
- *(dividends)* Add a delete method for dividends

## [1.27.1] - 2025-11-23

### 🐛 Bug Fixes

- *(dividends)* Ensure custom dividends are merged properly with official #129

### ⚙️ Miscellaneous Tasks

- Update lightweight charts to latest

## [1.27.0] - 2025-11-18

### 🚀 Features

- *(ui)* Add custom right axis and filterable timeline #127

## [1.26.0] - 2025-11-09

### 🚀 Features

- *(ui)* Adding trade should display fx rate used

### 🐛 Bug Fixes

- *(fx)* Always publish rates with respect to base currency

### 📚 Documentation

- Update README for nasdaq data source

## [1.25.0] - 2025-11-07

### 🚀 Features

- *(metrics)* Add new metrics cash flow view #124

### 🐛 Bug Fixes

- *(metrics)* Resolve memory leak when switching back from aggregated to granular metrics view

## [1.24.1] - 2025-11-03

### 🐛 Bug Fixes

- *(mdata)* Make dividends sg robust against malformed html #122

## [1.24.0] - 2025-11-01

### 🚀 Features

- *(blotter)* Add fallback for fx rates when inferring for non sgd trades #120

## [1.23.0] - 2025-10-25

### 🚀 Features

- *(dividends)* Add nasdaq to list of market data providers #118

## [1.22.1] - 2025-10-24

### 🐛 Bug Fixes

- *(ui)* Filters not recomputing aggregated changes

## [1.22.0] - 2025-10-24

### 🚀 Features

- *(ui)* Add ccy filters toggle group to positions page #115

## [1.21.0] - 2025-08-29

### 🚀 Features

- *(seed)* Add suntec reit and venture corp

### 🐛 Bug Fixes

- *(ui)* Blotter view should allow filtering on ticker name #113

## [1.20.1] - 2025-08-20

### 🐛 Bug Fixes

- *(ui)* Top 10 stocks were not reflecting most sold stocks

## [1.20.0] - 2025-08-18

### 🚀 Features

- *(blotter)* Add responsive view and ticker name display
- *(sgx)* Add top 10 fund flow and deprecate top 100

## [1.19.0] - 2025-08-08

### 🚀 Features

- *(portfolio)* Add a delete position endpoint and expose cli and mcp tools #107

### 🐛 Bug Fixes

- *(portfolio)* Separate out interfaces into portfolio reader and writer instead of getter

## [1.18.0] - 2025-08-07

### 🚀 Features

- *(ui)* Fix a memory leak with all dividends page

## [1.17.0] - 2025-07-23

### 🚀 Features

- *(user)* Implement user profile management feature

### 🐛 Bug Fixes

- Improve theme toggle button visibility in dark mode

## [1.16.0] - 2025-07-22

### 🚀 Features

- *(rdata)* Add seatrium to seed data
- *(hist,ui)* Add new handler to manually compute metrics and allow users to trigger in ui #95

### 🐛 Bug Fixes

- *(mdata)* Return empty list instead of null for yahoo dividends metadata #96

## [1.15.0] - 2025-07-21

### 🚀 Features

- *(mdata)* Allow fetching of benchmark interest rates
- *(mcp)* Add querying of benchmark rates into mcp server and modularize mcp inmplementation
- *(ui)* Add dashboard page to analytics for macro trends #92

## [1.14.0] - 2025-07-16

### 🚀 Features

- *(deps)* Upgrade vite from v5 to v7

### 🐛 Bug Fixes

- *(dividends)* Add retry to yahoo dividends endpoint

### ⚙️ Miscellaneous Tasks

- *(ai)* Add instructions to github copilot agent mode

## [1.13.0] - 2025-07-12

### 🚀 Features

- *(mcp)* Add mcp server #82
- *(analytics)* Make cron default schedule an env var as well #88

### 🐛 Bug Fixes

- *(portfolio)* Pass in empty book does not return all positions in GetPositions

## [1.12.0] - 2025-07-10

### 🚀 Features

- *(ui)* Add sector filters on 100 most traded reports #84

## [1.11.0] - 2025-07-09

### 🚀 Features

- *(analytics)* Add sector funds flow parsing backend handlers and ui component #79

## [1.10.0] - 2025-07-04

### 🚀 Features

- *(hist)* Add new endpoints to create delete and list custom metrics jobs #72
- *(ui)* Allow metrics to respect book filters #72
- *(hist)* Add book filters to import and export csv #72
- *(hist)* Allow import and export historical metrics csv with book filters #72

### ⚙️ Miscellaneous Tasks

- *(seed)* Add st and sia engineering to seed

## [1.9.1] - 2025-07-04

### 🐛 Bug Fixes

- *(ui)* Add report name shortening functionality and testing instructions

## [1.9.0] - 2025-07-03

### 🚀 Features

- *(metrics)* Enhance metrics to allow filtering by book #72
- *(metrics)* Add book label to metrics
- *(hist)* Refactor interfaces to smaller components
- *(hist)* Allow optional book_filter parameter when storing metrics
- *(hist)* Allow optional book_filter parameters when getting metrics
- *(ui)* Add optionality to plot pnl against irr #75
- *(ui)* Add p&l to the left bottom of the screen for mobile view
- *(ui)* Add all dividends view #73

### 🐛 Bug Fixes

- *(hist)* Unschedule was not stopping all scheduled tasks
- *(ui)* Handle sgx report naming change and add unit tests #74

### ⚙️ Miscellaneous Tasks

- *(ui)* Update mantine to v7.17.8
- *(metrics)* Add todos on metric by book

## [1.8.0] - 2025-06-25

### 🚀 Features

- *(ui)* Add pnl column to historical metrics
- *(ui)* Improve mobile ux experience by auto collapsing navbar on click

### 🐛 Bug Fixes

- *(ui)* Positions and dividends throwing error due to incorrect aggregation

### 📚 Documentation

- Minor tweak

## [1.7.0] - 2025-06-19

### 🚀 Features

- Rename trader to book across portfolio and blotter
- *(migrations)* Create new migrations struct to handle migrations
- *(migrations)* Add step to migrate portfolio positions in v1.7.0
- *(rdata)* Add uol to seed
- *(migrations)* Wire up to always run when entrypoint is called
- *(blotter)* Export csv should postfix yyyymmdd
- *(ui)* Allow granular view of positions by book #62

### 🐛 Bug Fixes

- *(blotter)* Updating trades didn't persist to db #65

### ⚙️ Miscellaneous Tasks

- *(migrations)* Small refactor in docs

## [1.6.2] - 2025-06-10

### 🐛 Bug Fixes

- *(refdata)* Handle missing id when adding new ref data
- *(refdata)* Update redux store when upserting ref data
- *(ui)* Add a loading indicator to positions table
- *(pipelines)* Revert action-gh-release to an older version

## [1.6.1] - 2025-06-07

### 🐛 Bug Fixes

- *(ui)* Add sector to most traded stocks report

## [1.6.0] - 2025-06-07

### 🚀 Features

- *(analytics)* Concatenate most commonly traded stocks across multiple funds flow report
- *(ui)* Add heat map of most traded stocks #59

### 🐛 Bug Fixes

- *(sgx)* Increase query limits on sgx reports #57

## [1.5.0] - 2025-06-06

### 🚀 Features

- *(ai)* List all sgx reports and store ai analysis in level db
- *(ai)* Only download sgx reports if not processed previously
- *(ai)* Add handler endpoint to fetch all previously stored ai analysis
- *(historical)* Schedule the collection of sgx funds flow report and analysis
- *(analytics)* Refactor process report into download and analyze report steps
- *(ai)* Create endpoints to download sgx reports #57
- *(ui)* Add new analytics reports component to view sgx reports
- *(ai)* Allow analyzing multiple reports
- *(ui)* Add functionality to run gemini analysis on downloaded reports
- *(ai)* Fetch gemini api key via environment variable

### 🐛 Bug Fixes

- *(ai)* Remove context as input parameter

## [1.4.2] - 2025-06-03

### 🐛 Bug Fixes

- *(mdata)* Mocking browser and setting content type to gzip means handler should cater for gzipped responses #53

## [1.4.1] - 2025-06-03

### 🐛 Bug Fixes

- *(ui)* UI was passing in wrong date when fetching non sgd fx rate in add_blotter due to UTC conversion #53

## [1.4.0] - 2025-06-03

### 🚀 Features

- *(ai)* Add ai analytics tool #50
- *(ai)* Update ai funds flow prompt to be more relevant to trading #50
- *(ai)* Switch default model #50

### 🐛 Bug Fixes

- *(mdata)* Rate limit yahoo finance historical queries and impersonate chrome #53
- *(release)* Pipeline was uploading full changelog for each release instead of deltas

## [1.3.0] - 2025-06-02

### 🚀 Features

- *(go)* Update go from 1.23 to 1.24 #47

### 🐛 Bug Fixes

- Set fx rate to 1 for SG Govies when base ccy is SGD #51
- Handle 404 errors on all endpoints other than / root #17

## [1.2.2] - 2025-05-23

### 🐛 Bug Fixes

- *(refdata)* Fix bug with refdata upsert not working in ui #48

## [1.2.1] - 2025-05-18

### 🐛 Bug Fixes

- Handle bug with empty historical metrics #42
- Delete multiple historical metrics at one go #43
- Scheduler needs to be started for it to process tasks #44

### ⚙️ Miscellaneous Tasks

- Add bugs and issues template

## [1.2.0] - 2025-05-17

### 🚀 Features

- *(historical)* Add CRUD functionality and new endpoints #36
- *(ui)* Allow navigation from positions to dividends #40
- *(ui)* Add a historical metrics component #38
- *(ui)* Add lightweight charts to plot historical metrics #26

### 📚 Documentation

- Add a version badge

### ⚙️ Miscellaneous Tasks

- Add issues template

## [1.1.2] - 2025-05-12

### 🐛 Bug Fixes

- Merge pull request #37 from rahulbhataniya/Issue35_show_entire_history_of_changelog

### 📚 Documentation

- Add CONTRIBUTING guidelines

### ⚙️ Miscellaneous Tasks

- Only run release when it is a non fork PR
- Refactor pipelines into a reusable release workflow

## [1.1.1] - 2025-05-10

### 🐛 Bug Fixes

- *(ui)* Add blotter trades errors when ccy is sgd

### 📚 Documentation

- *(ui)* Refresh documentations for ui

## [1.1.0] - 2025-05-09

### 🚀 Features

- *(dividends)* Allow fetching of all dividends
- *(dividends)* Add caching when fetching all dividends
- *(ui)* Create component for aggregated dividends by year
- *(ui)* Bring in blotter trades to compute net purchases and sales per year
- *(blotter)* Add fx property to each trade
- *(mdata)* Add handler to extract historical market data
- *(config)* Add base ccy to config
- *(fxinfer)* Add new endpoint to infer blotter trades fx
- *(blotter)* Add backwards compatibility for files with no fx column
- *(ui)* Add fx column to blotter table
- *(ui)* Add fx inferring to blotter trades in settings drawer
- *(fxinfer)* Add new endpoint to fetch relevant current fx rates based on blotter trades
- *(ui)* Dividends are revalued at current rates and add portfolio stats excl gov
- *(ui)* Infer historical fx when upserting a trade with 0 fx rate
- *(metrics)* Add metrics service to compute irr
- *(scheduler)* Add a scheduler module to create jobs
- *(scheduler)* Switch to cron scheduler for flexibility
- *(config)* Add nesting to config yaml to make it clearer
- *(historical)* Redo getMetrics to fetch all data
- *(historical)* Add new handler to get all historical metrics
- *(historical)* Add configuration to start historical metrics scheduling when app starts
- *(versioning)* Introduce git cliff for semantic versioning

### 🐛 Bug Fixes

- *(ui)* Fx rate needs to populate when updating a trade in blotter

### 📚 Documentation

- Update README and shift roadmap to github issues

### ⚙️ Miscellaneous Tasks

- *(versioning)* Redo versioning
- *(versioning)* Only allow releases on PRs to main

## [1.0.10] - 2025-04-26

### 🚀 Features

- *(ui)* Blotter table should show most recent trades first
- *(ui)* Implement export blotter button in ui and fix missing status column in backend

## [1.0.9] - 2025-04-25

### 🐛 Bug Fixes

- *(ui)* Allow decimals in qty and persist state when routing to add trade from blotter view

## [1.0.8] - 2025-04-03

### 🐛 Bug Fixes

- Correct status validation on ui and extra leading slash in post method

## [1.0.7] - 2025-03-29

### 🚀 Features

- Enrich position with market data on 2 threads

### 📚 Documentation

- Update roadmap

## [1.0.6] - 2025-03-28

### 🚀 Features

- Allow users to navigate to blotter from positions component in ui

### 🐛 Bug Fixes

- Fcot final dividends on wrong date

### ⚙️ Miscellaneous Tasks

- Automatically generate release notes

## [1.0.5] - 2025-03-28

### 🐛 Bug Fixes

- Opening and closing position twice leads to incorrect pnl
- Ui blanks out when invalid ticker is imported into blotter

### 🧪 Testing

- Add test on enrich position for pnl calc on closed positions

## [1.0.4] - 2025-03-25

### 🚀 Features

- Allow 0 price trades

## [1.0.3] - 2025-03-23

### 🚀 Features

- Add clean up setting to reset blotter and positions

### 🐛 Bug Fixes

- Lxc not correctly storing update command

## [1.0.2] - 2025-03-23

### 🚀 Features

- Add settings component for auto expiries

## [1.0.1] - 2025-03-22

### 🚀 Features

- Print version during lxc installation and fix update script

### 🐛 Bug Fixes

- Unable to delete when all rows are selected

### 📚 Documentation

- Add section on developer

### ⚙️ Miscellaneous Tasks

- Set up automatic version bumping and release on merge to main
- Create tags when version is bumped

## [1.0.0] - 2025-03-22

### 🚀 Features

- First commit
- Update dependencies
- Tidy up start up
- Add http server
- Add blotter service
- Wire up db and blotter in application
- Allow configurable db path
- Update cash and bond asset classes
- Add new getters for blotter
- Refactor blotter and add tests
- Server should compose a blotter service
- Remove dataframes from blotter service
- Add handler to add trade in blotter
- Add sequence number to blotter
- Add broker to blotter
- Add event subscription functionality to blotter
- Add portfolio service
- Add portfolio tests
- Wire up portfolio service to application
- Add clean db functionality in makefile
- Wire portfolio to subscribe to blotter and add more validation
- Update average price upon new trades
- Add market data mdata package
- Add dividends.sg mdata source
- Register mdata handlers in server
- Mdata dividends should be sorted
- Wire in dividends sg to handlers
- Introduce ticker reference data and shift asset class there
- Implement ref data manager
- Make configurations a cli flag
- Implement import from csv for blotter trades
- Implement export to csv for blotter trades
- Add short name to ref data
- Add account to blotter
- Seed more ticker references
- Add swagger documentation
- Enhance get position to return reference data
- Wire up dividends manager
- Convert config to singleton
- Add dividends witholding tax capability
- Add country of domicile for ticker reference for withholding tax computation
- Add ireland withholding tax configuration
- Extract out mdata interface
- Extract interfaces from blotter
- Add dividends computation tests and add ticker to dividends metadata
- Cache dividends instead of hitting endpoint multiple times
- Refactor mdata to use rdata for data source routing
- Add pnl computation logic to portfolio
- Add dividends to pnl computation logic in portfolio
- Cache fetching of stock price for yahoo data source
- Support singapore savings bond in portfolio
- Ticker reference to support crypto
- Implement get asset price for ssb
- Seed fx ticker references
- Compute dividends for ssb
- Support non sgd dividends
- Support mas bills dividends and asset price
- Add web-ui skeleton
- Add more stocks to ref data seed
- Update web ui dependencies
- Edit styling on ui
- Add navigation controller
- Add navigation logic to route to add trades
- Add cors middleware and remove fetch blotter hardcode
- Create a basic form to add blotter trades
- Add redux store to fetch ref data
- Add ref data type
- Add autocomplete for ticker in add trade and set defaults for trader and broker on submit
- Implement react-router-dom instead of custom controller
- Add trade date to blotter table
- Add submit post call to adding a trade to blotter from ui
- Upgrade react router dom to v7.1.3
- Allow row selection for blotter table
- Allow deletion and update of blotter trades
- Add functionality to export reference data as yaml
- Add custom action for blotter toolbar to delete and add trades
- Allow for user to specify trade value instead of only trade price
- Allow for deletion and adding of trades via ui
- Allow updating of trades
- Add reference data fetch ui component
- Update ref data svc to handle updates, addition and deletion
- Cosmetic change on blotter navbar
- Add update functionality to reference data
- Default to show column filters
- Add skeleton for web ui building
- Wire up web ui via conditional build tags
- Add position table in ui
- Aggregate ssb and tbill positions
- Add ticker name skeleton
- Add totals to positions
- Add breakdown for non government bonds
- Add ticker name to positions
- Positions endpoint should return current price as well
- Add percentage relative to total portfolio in positions
- Add original trade id and trade lifecycle to blotter trades
- Add autoclosing of expired trades functionality to portfolio service
- Add autoclosing of expired positions functionality
- Revalue position against sgd
- Add lxc skeleton support
- Embed default config and ref data yaml files
- Point lxc script to dl from github
- Abstract commands to compile UI into the backend
- Update api calls to use window.location instead of hardcoded ones
- Allow collapsible navbar on small screens
- Add upload functionality in ui for blotter trades
- Add dividends breakdown to ui component
- Allow for user supplied custom dividends metadata
- Implement GetAssetPrice for dividends sg
- Add vt to ref data seed
- Allow custom dividends upload by csv
- Refactor /mdata/dividend to /mdata/dividends
- Update navbar sizing
- Pick latest version dynamically for lxc installation

### 🐛 Bug Fixes

- Handle bug with sell trades in portfolio positions
- Handle divison by zero error
- Compute dividends had issues with sorting
- Dividends handler prematurely exiting
- Future dividends should not be included in positions computation
- Zerorize dividends since we recompute all the time
- Trigger dispatch without storing ref data as a variable
- Typo with reference data nav bar
- Handle errors correctly when submitting trades
- Add account to submit trade form
- Auto compute price when value is specified
- Bug with update ref data not updating right field
- Correct content-type for static files
- Fix failing tests
- Invalid mv when position closed and update maturity date for bills
- Send positions back in base currency
- Yahoo finance dividends caching was colliding with asset price
- Default to bind to all addresses and listen on 8080
- Development ui should point to hardcoded backend
- Portfolio endpoints should not return null but empty arrays
- Expose dividends by portfolio as get instead of post
- Dividends should minimally be 4 decimals
- Dividends sg was not caching dividends metadata correctly
- Dividends metadata for sg bonds refering to wrong columns
- Invalid import message

### 🚜 Refactor

- Portfolio should not instantiate mdata and rdata
- Mdata dividends renamed to dividends metadata
- Shift reference data from internal into a package and extract as interface
- Refactor mdata GetStockPrice to GetAssetPrice
- Shift all api handlers to /api/v1 prefix
- Shift Blotter and Controller to components

### 📚 Documentation

- Add sample curl commands to README
- Add product roadmap
- Add ui images
- Update roadmap implementations

### 🧪 Testing

- Add more tests for dividends handler

### ⚙️ Miscellaneous Tasks

- Update README
- Update README with project structure
- Add workflow to run automated tests
- Add release workflow dispatch
- Use action-gh-release for release

<!-- generated by git-cliff -->
