import fs from 'fs'
import React from 'react'

import { HeimNav, MainPage, Markdown } from './common'

// The text is templated using Go templates at runtime, hence the {{...}} syntax here and in libpage.md.
module.exports = (
  <MainPage title="Euphoria: Open Source: {{.Name}}" meta={<meta name="go-import" content="{{.GoImport}}" />} nav={<HeimNav />} noFooter>
    <Markdown className="text-page libpage" content={fs.readFileSync(__dirname + '/libpage.md', 'utf8')} />
  </MainPage>
)
