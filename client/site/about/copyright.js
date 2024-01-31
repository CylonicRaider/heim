import fs from 'fs'
import React from 'react'

import { MainPage, PolicyNav, Markdown } from '../common'

module.exports = (
  <MainPage title="Euphoria: Copyright Policy" nav={<PolicyNav selected="copyright" />}>
    <Markdown className="text-page policy" content={fs.readFileSync(__dirname + '/copyright.md', 'utf8')} />
  </MainPage>
)
