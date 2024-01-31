import fs from 'fs'
import React from 'react'

import { MainPage, PolicyNav, Markdown } from '../common'

module.exports = (
  <MainPage title="Euphoria: Terms of Service" nav={<PolicyNav selected="terms" />}>
    <Markdown className="text-page policy" content={fs.readFileSync(__dirname + '/terms.md', 'utf8')} />
  </MainPage>
)
