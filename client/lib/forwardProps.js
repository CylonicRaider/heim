import _ from 'lodash'

export default function forwardProps(self) {
  // TODO: check for unexpected props being swallowed
  return _.pick(self.props, ['id', 'className', 'title'])
}
