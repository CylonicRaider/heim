import React from 'react'
import createReactClass from 'create-react-class'
import PropTypes from 'prop-types'
import Immutable from 'immutable'

import accountAuthFlow from '../stores/accountAuthFlow'
import FastButton from './FastButton'

export default createReactClass({
  displayName: 'AccountButton',

  propTypes: {
    account: PropTypes.instanceOf(Immutable.Map),
    onOpenAccountAuthDialog: PropTypes.func,
    onOpenAccountSettingsDialog: PropTypes.func,
  },

  mixins: [require('react-immutable-render-mixin')],

  render() {
    if (this.props.account) {
      const name = this.props.account.get('name')
      const email = this.props.account.get('email')
      return (
        <FastButton className="account-button signed-in" onClick={this.props.onOpenAccountSettingsDialog}>
          <div className="account-info">
            <div className="status">signed in</div>
            <div className="name" title={email}>{name || email}</div>
          </div>
        </FastButton>
      )
    }

    const title = accountAuthFlow.registerAvailable ? 'sign in or register' : 'sign in'
    return <FastButton className="account-button" onClick={this.props.onOpenAccountAuthDialog}>{title}</FastButton>
  },
})
