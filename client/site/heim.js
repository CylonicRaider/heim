import fs from 'fs'
import React from 'react'

import { MainPage, Markdown, links } from './common'

module.exports = (
  <MainPage title="Euphoria: Open Source">
    <Markdown className="text-page opensource" content={fs.readFileSync(__dirname + '/heim.md', 'utf8')} />
  </MainPage>
)
