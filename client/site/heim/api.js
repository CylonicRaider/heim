import fs from 'fs'
import React from 'react'

import { MainPage, HeimNav, Markdown } from '../common'

module.exports = (
  <MainPage title="Euphoria: API" nav={<HeimNav selected="api" />} sidebar={true}>
    <Markdown className="text-page api" content={fs.readFileSync(__dirname + '/../../../doc/api.md', 'utf8')} />
  </MainPage>
)
