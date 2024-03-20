import fs from 'fs'
import React from 'react'

import { HeimNav, MainPage, Markdown } from './common'

module.exports = (
  <MainPage title="Euphoria: Open Source" nav={<HeimNav selected="opensource" />}>
    <Markdown className="text-page opensource" content={fs.readFileSync(__dirname + '/heim.md', 'utf8')} />
  </MainPage>
)
