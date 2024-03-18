/* eslint-disable arrow-body-style */
import '@babel/polyfill'

import _ from 'lodash'
import merge from 'merge-stream'
import gulp from 'gulp'
import gfile from 'gulp-file'
import gzip from 'gulp-gzip'
import gtemplate from 'gulp-template'
import less from 'gulp-less'
import autoprefixer from 'gulp-autoprefixer'
import uglify from 'gulp-uglify'
import sourcemaps from 'gulp-sourcemaps'
import source from 'vinyl-source-stream'
import buffer from 'vinyl-buffer'
import watchify from 'watchify'
import browserify from 'browserify'
import envify from 'envify/custom'
import serve from 'gulp-serve'
import fs from 'fs'
import ReactHTMLEmail from 'react-html-email'
import { exec } from 'child_process'
import colors from 'ansi-colors'
import log from 'fancy-log'
import through2 from 'through2'
import babelify from 'babelify'
import brfs from 'brfs'
import colorSupport from 'color-support'
import curCommit from 'current-commit'

let watching = false
const heimDest = './build/heim'
const heimStaticDest = heimDest + '/static'
const heimPagesDest = heimDest + '/pages'
const embedDest = './build/embed'
const emailDest = './build/email'

const production = process.env.NODE_ENV === 'production'

const heimOptions = {
  HEIM_ORIGIN: process.env.HEIM_ORIGIN,
  HEIM_PREFIX: process.env.HEIM_PREFIX || '',
  EMBED_ORIGIN: process.env.EMBED_ORIGIN,
  NODE_ENV: process.env.NODE_ENV,
  ACCOUNTS_DISABLED_NOTE: process.env.ACCOUNTS_DISABLED_NOTE,
}

// apply an ansi-colors function if the terminal supports color
function doColor(color, string) {
  return colorSupport.hasBasic ? color(string) : string
}

// more meaningful name
function noop() {
  return through2.obj()
}

// allow disabling this one on weak machines
function heimUglify() {
  if (!production || process.env.HEIM_NO_UGLIFY) return noop()
  return uglify()
}

// via https://github.com/tblobaum/git-rev
function shell(cmd, cb) {
  exec(cmd, { cwd: __dirname }, (err, stdout) => {
    if (err) {
      throw err
    }
    cb(stdout.trim())
  })
}

// FIXME: replace with a more robust js loader
function reload(moduleName) {
  delete require.cache[require.resolve(moduleName)]
  return require(moduleName)  // eslint-disable-line import/no-dynamic-require
}

function handleError(title) {
  return function handler(err) {
    log(colors.red(title + ':'), err.message)
    if (watching) {
      this.emit('end')
    } else {
      process.exit(1)
    }
  }
}

function heimBrowserify(files, args) {
  return browserify(files, args)
    .transform(babelify, {presets: ['@babel/preset-env',
      '@babel/preset-react']})
    .transform(brfs)
}

function heimBundler(args) {
  return heimBrowserify('./lib/client.js', args)
    .transform(envify(heimOptions))
}

function embedBundler(args) {
  return heimBrowserify('./lib/embed.js', args)
    .transform(envify({
      HEIM_ORIGIN: process.env.HEIM_ORIGIN,
    }))
}

gulp.task('heim-git-commit', (done) => {
  curCommit('..', (err, gitRev) => {
    if (gitRev) {
      log('Git commit:', doColor(colors.bold, gitRev))
    } else {
      log('Git commit: N/A')
    }
    process.env.HEIM_GIT_COMMIT = heimOptions.HEIM_GIT_COMMIT = gitRev
    done()
  })
})

gulp.task('heim-js', ['heim-git-commit', 'heim-less'], () => {
  return heimBundler({debug: true})
    // share some libraries with the global namespace
    // doing this here because these exposes trip up watchify atm
    .require('lodash', {expose: 'lodash'})
    .require('react', {expose: 'react'})
    .require('reflux', {expose: 'reflux'})
    .require('immutable', {expose: 'immutable'})
    .require('moment', {expose: 'moment'})
    .require('querystring', {expose: 'querystring'})
    .bundle()
    .pipe(source('main.js'))
    .pipe(buffer())
    .pipe(sourcemaps.init({loadMaps: true}))
    .pipe(heimUglify())
    .pipe(sourcemaps.write('./', {includeContent: true}))
    .on('error', handleError('heim browserify error'))
    .pipe(gulp.dest(heimStaticDest))
    .pipe(gzip())
    .pipe(gulp.dest(heimStaticDest))
})

gulp.task('fast-touch-js', () => {
  return gulp.src('./site/lib/fast-touch.js')
    .pipe(heimUglify())
    .on('error', handleError('fastTouch browserify error'))
    .pipe(gulp.dest(heimStaticDest))
})

gulp.task('embed-js', () => {
  return embedBundler({debug: true})
    .bundle()
    .pipe(source('embed.js'))
    .pipe(buffer())
    .pipe(sourcemaps.init({loadMaps: true}))
    .pipe(heimUglify())
    .pipe(sourcemaps.write('./', {includeContent: true}))
    .on('error', handleError('embed browserify error'))
    .pipe(gulp.dest(embedDest))
    .pipe(gzip())
    .pipe(gulp.dest(embedDest))
})

gulp.task('raven-js', ['heim-git-commit', 'heim-js'], (done) => {
  shell('sha256sum build/heim/static/main.js | cut -d " " -f 1', (releaseHash) => {
    if (releaseHash) {
      log('Release hash:', doColor(colors.bold, releaseHash))
    } else {
      log('Release hash: N/A')
    }
    heimBrowserify('./lib/raven.js')
      .transform(envify(_.extend({
        SENTRY_ENDPOINT: process.env.SENTRY_ENDPOINT,
        HEIM_RELEASE: releaseHash,
      }, heimOptions)))
      .bundle()
      .pipe(source('raven.js'))
      .pipe(buffer())
      .pipe(heimUglify())
      .on('error', handleError('raven browserify error'))
      .pipe(gulp.dest(heimStaticDest))
      .pipe(gzip())
      .pipe(gulp.dest(heimStaticDest))
      .on('end', done)
  })
})

gulp.task('heim-less', () => {
  return gulp.src(['./lib/main.less', './lib/crashed.less', './lib/od.less', './lib/gadgets/*.less', './site/*.less'])
    .pipe(less({compress: true, math: 'always'}))
    .on('error', handleError('LESS error'))
    .pipe(autoprefixer({cascade: false}))
    .on('error', handleError('autoprefixer error'))
    .pipe(gulp.dest(heimStaticDest))
    .pipe(gzip())
    .pipe(gulp.dest(heimStaticDest))
})

gulp.task('emoji-static', () => {
  const emoji = require('./lib/emoji').default
  const twemojiPath = 'node_modules/.resources/emoji-svg/'
  const leadingZeroes = /^0*/
  const emojiFiles = _.map(fs.readdirSync(twemojiPath), (p) => {
    const m = /^([0-9a-f-]+)\.svg$/.exec(p)
    if (!m) {
      return null
    }
    return m[1]
  })
  const emojiCodes = _.uniq(_.concat(emoji.codes, emojiFiles))
  const lessSource = _.map(_.compact(emojiCodes), (code) => {
    const twemojiName = code.replace(leadingZeroes, '')
    let emojiPath = './res/emoji/' + twemojiName + '.svg'
    if (!fs.existsSync(emojiPath)) {
      emojiPath = twemojiPath + twemojiName + '.svg'
    }
    if (!fs.existsSync(emojiPath)) {
      return ''
    }
    return '.emoji-' + code + ' { background-image: data-uri("' + emojiPath + '") }'
  }).join('\n')

  const lessFile = gfile('emoji.less', lessSource, {src: true})
    .pipe(less({compress: true}))
    .pipe(gulp.dest(heimStaticDest))
    .pipe(gzip())
    .pipe(gulp.dest(heimStaticDest))

  const indexFile = gfile('emoji.json', JSON.stringify(emoji.index), {src: true})
    .pipe(gulp.dest(heimDest))

  return merge([lessFile, indexFile])
})

gulp.task('heim-static', () => {
  return gulp.src(['./static/**/*', './site/static/**'])
    .pipe(gulp.dest(heimStaticDest))
})

gulp.task('embed-static', () => {
  return gulp.src('./static/robots.txt')
    .pipe(gulp.dest(embedDest))
})

gulp.task('heim-html', ['heim-git-commit'], () => {
  return gulp.src(['./lib/room.html'])
    .pipe(gtemplate(heimOptions))
    .pipe(gulp.dest(heimPagesDest))
})

gulp.task('embed-html', () => {
  return gulp.src(['./lib/embed.html'])
    .pipe(gulp.dest(embedDest))
})

gulp.task('site-templates', ['heim-git-commit'], () => {
  const page = reload('./site/page.js')
  const pages = [
    'home',
    'error',
    'verify-email',
    'reset-password',
    'about',
    'about/values',
    'about/conduct',
    'about/hosts',
    'about/terms',
    'about/privacy',
    'about/copyright',
    'heim',
    'heim/api',
  ]

  return merge(_.map(pages, (name) => {
    const html = page.render(reload('./site/' + name))
    return gfile(name + '.html', html, {src: true})
  }))
    .pipe(gulp.dest(heimPagesDest))
})

gulp.task('email-templates', () => {
  ReactHTMLEmail.injectReactEmailAttributes()
  ReactHTMLEmail.configStyleValidator({platforms: ['gmail']})
  const renderEmail = require('react-html-email').renderEmail
  const emails = ['welcome', 'room-invitation', 'room-invitation-welcome', 'verification', 'password-changed', 'password-reset']

  const htmls = merge(_.map(emails, (name) => {
    const html = renderEmail(reload('./emails/' + name))
    return gfile(name + '.html', html, {src: true})
  }))

  const txtCommon = reload('./emails/common-txt.js').default
  const txts = merge(_.map(emails, (name) => {
    return gulp.src('./emails/' + name + '.txt')
      .pipe(gtemplate(txtCommon))
  }))

  return merge(htmls, txts)
    .pipe(gulp.dest(emailDest))
})

gulp.task('email-hdrs', () => {
  return gulp.src('./emails/*.hdr').pipe(gulp.dest(emailDest))
})

gulp.task('email-static', () => {
  return gulp.src('./emails/static/*.png').pipe(gulp.dest(emailDest + '/static'))
})

function watchifyTask(name, bundler, outFile, dest) {
  gulp.task(name, ['build-statics'], () => {
    // via https://github.com/gulpjs/gulp/blob/master/docs/recipes/fast-browserify-builds-with-watchify.md
    const watchBundler = watchify(bundler({debug: true, ...watchify.args}))

    function rebundle() {
      return watchBundler.bundle()
        .on('error', handleError('JS (' + name + ') error'))
        .pipe(source(outFile))
        .pipe(gulp.dest(dest))
        .pipe(gzip())
        .pipe(gulp.dest(dest))
    }

    watchBundler.on('log', log.bind(null, doColor(colors.green, 'JS (' + name + ')')))
    watchBundler.on('update', rebundle)
    return rebundle()
  })
}

watchifyTask('heim-watchify', heimBundler, 'main.js', heimStaticDest)
watchifyTask('embed-watchify', embedBundler, 'embed.js', embedDest)

gulp.task('build-emails', ['email-templates', 'email-hdrs', 'email-static'])
gulp.task('build-statics', ['raven-js', 'fast-touch-js', 'heim-less', 'emoji-static', 'heim-static', 'embed-static', 'heim-html', 'embed-html', 'site-templates'])
gulp.task('build-browserify', ['heim-js', 'embed-js'])

gulp.task('watch', () => {
  watching = true
  gulp.watch('./lib/**/*.less', ['heim-less'])
  gulp.watch('./res/**/*', ['heim-less', 'emoji-static'])
  gulp.watch('./site/**/*.less', ['heim-less'])
  gulp.watch('./lib/**/*.html', ['heim-html', 'embed-html'])
  gulp.watch('./static/**/*', ['heim-static', 'embed-static'])
  gulp.watch('./site/**/*', ['site-templates'])
  gulp.watch('./emails/*', ['email-templates'])
  gulp.watch('./emails/static/*', ['email-static'])
})

gulp.task('build', ['build-statics', 'build-browserify', 'build-emails'])
gulp.task('default', ['build-statics', 'build-emails', 'watch', 'heim-watchify', 'embed-watchify'])

gulp.task('serve-heim', serve({
  hostname: '0.0.0.0',
  port: 8080,
  root: heimDest,
  middleware: function serveHeim(req, res, next) {
    if (req.url === '/') {
      req.url = '/pages/home.html'
    } else if (/^\/room\/\w+\/?/.test(req.url)) {
      req.url = '/pages/room.html'
    } else if (!/^\/static/.test(req.url)) {
      req.url = '/pages' + req.url + '.html'
    }
    next()
  },
}))

gulp.task('serve-embed', serve({
  hostname: '0.0.0.0',
  port: 8081,
  root: embedDest,
  middleware: function serveEmbed(req, res, next) {
    req.url = req.url.replace(/^\/\?\S+/, '/embed.html')
    next()
  },
}))

gulp.task('develop', ['serve-heim', 'serve-embed', 'default'])
