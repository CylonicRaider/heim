import fs from 'fs'
import React from 'react'

import { HeimNav, MainPage, Markdown, links } from './common'

// One could argue that the VCS type (git) belongs to where the repository is defined rather than here. However, some
// other code in the backend hard-codes a dependency on Git as well. As long as it's found by a full-text search for
// "git" in the unlikely case that this project will ever transition to another VCS, it's fine.
const goImport = links.heimGoPackage + ' git ' + links.heimSourceRepo + ' server'

module.exports = (
  <MainPage title="Euphoria: Open Source" meta={<meta name="go-import" content={goImport} />} nav={<HeimNav selected="opensource" />}>
    <Markdown className="text-page opensource" content={fs.readFileSync(__dirname + '/heim.md', 'utf8')} />
  </MainPage>
)
