import React from 'react'

import { MainPage } from './common'

module.exports = (
  <MainPage title="Euphoria: Error" className="error-page">
    <h1>Error</h1>
    {'{{if .Code}}'}<h2>{'{{.Code}}'}</h2>{'{{end}}'}
    <h3>{'{{.Message}}'}</h3>
  </MainPage>
)
