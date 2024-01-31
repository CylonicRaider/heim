import fs from 'fs'
import React from 'react'

import { MainPage, PolicyNav, Markdown } from '../common'

module.exports = (
  <MainPage title="Euphoria: Code of Conduct" nav={<PolicyNav selected="conduct" />}>
    <Markdown className="text-page policy" content={fs.readFileSync(__dirname + '/conduct.md', 'utf8')} />
  </MainPage>
)
