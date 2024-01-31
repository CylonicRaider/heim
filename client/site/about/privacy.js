import fs from 'fs'
import React from 'react'

import { MainPage, PolicyNav, Markdown } from '../common'

module.exports = (
  <MainPage title="Euphoria: Privacy Policy" nav={<PolicyNav selected="privacy" />}>
    <Markdown className="text-page policy" content={fs.readFileSync(__dirname + '/privacy.md', 'utf8')} />
  </MainPage>
)
