import assert from 'assert'

import actions from '../lib/heim/actions'

describe('global actions', () => {
  describe('setup action', () => {
    it('should be synchronous', () => {
      assert.equal(actions.setup.sync, true)
    })
  })
})
