import fs from 'fs'
import React from 'react'

import { MainPage, PolicyNav, Markdown } from '../common'

module.exports = (
  <MainPage title="euphoria: api" nav={<PolicyNav selected="api" />}>
    <Markdown className="text-page api" content={fs.readFileSync(__dirname + '/../../../doc/api.md', 'utf8')} />
  </MainPage>
)
