var pkg = require('./package.json');

var gulp = require('gulp');
var fs = require('fs');
var path = require('path');

var rimraf       = require('rimraf');
var git          = require('git-rev')

//var browserSync  = require('browser-sync');
//var browserify   = require('browserify');
//var webpack      = require('webpack');
//var watchify     = require('watchify');
//var source       = require('vinyl-source-stream');
//var changed      = require('gulp-changed');
//var imagemin     = require('gulp-imagemin');
var sass         = require('gulp-sass');
//var sass         = require('gulp-ruby-sass');

var templateCache = require('gulp-angular-templatecache');
var ngAnnotate   = require('gulp-ng-annotate');
var header       = require('gulp-header');
//var notify       = require('gulp-notify');
var prefix       = require('gulp-autoprefixer');
var inject       = require('gulp-inject');
var gutil        = require('gulp-util');
var prettyHrtime = require('pretty-hrtime');
var startTime;

var headerFileContent = fs.readFileSync('src/header.txt', 'utf-8')

// JS:
//  * take all my .js files, browserify them
//    * ng-min the angular stuff
//      * https://www.npmjs.org/package/browserify-ngmin
//    * I get a single output
//    * enable source maps in DEV
//    * inject in app.js section
//  * take vendor .js files
//    * min those who aren't min.js
//    * concat the rest
//    * inject in vendor.js section
//  * compile all the templates
//    * inject in templates:js

//  * take .scss files anywhere
//    * sass compile them
//    * inject back a single .css
//      * enable source maps in DEV
//
//  * imagemin
//    * only those modified
//
// Reads from static/*
// Outputs to:
// TODO:
// * Add headers and footers to output, ideally including Git revision, branch, copyright notice.
//    * gulp-rev https://www.npmjs.org/package/gulp-rev
// * Add gulp-sass, minifyCSS, source maps
// * Add assets versioning
// * Add AngularJS stuff
//    * fill in Template Cache
//    * angular-injector fixes
//    * browserify deals with the rest
//      * watchify: https://gist.github.com/crisward/9342850
//    * Inject in the sections of the HTML: http://netengine.com.au/blog/gulp-and-angularjs-a-love-story-or-the-old-wheel-was-terrible-check-out-my-new-wheel/
// * Add UglifyJS
//   * With source maps
// * Add bundling out of the index.html file

// JS & Angular

gulp.task('app', ['git-rev', 'clean:app', 'template_cache'], function() {
    return gulp.src('src/js/*.js')
        .pipe(ngAnnotate())
        .pipe(headerPipe())
        .pipe(gulp.dest('static/js'))
});

gulp.task('clean:app', function(cb) {
    rimraf("static/js", cb);
});

gulp.task('template_cache', ['git-rev'], function() {
    return gulp.src('src/tpl/**/*.html')
        .pipe(templateCache({module: "plotbot"}))
        .pipe(headerPipe())
        .pipe(gulp.dest('static/js'))
});

// Index page

gulp.task('build', ['copy:vendor', 'app', 'sass'], function() {
    return injectIndexHtml();
});

gulp.task('index', function() {
    return injectIndexHtml();
})

function injectIndexHtml() {
    var targetVendor = gulp.src('static/vendor/*.js', {read: false});
    var targetAppJs = gulp.src('static/js/*.js', {read: false});
    var targetCss = gulp.src('static/css/*.css', {read: false});
    return gulp.src('src/index.html')
        .pipe(inject(targetVendor, {starttag: '<!-- inject:vendor:{{ext}} -->'}))
        .pipe(inject(targetAppJs, {starttag: '<!-- inject:app:{{ext}} -->'}))
        .pipe(inject(targetCss))
        .pipe(gulp.dest('static'));
}

gulp.task('copy:vendor', ['clean:vendor'], function() {
    return gulp.src('src/vendor/*.js')
        .pipe(gulp.dest('static/vendor'));
});

gulp.task('clean:vendor', function(cb) {
    rimraf("/static/vendor", cb)
});


// CSS

gulp.task('sass', ['git-rev'], function() {
    return gulp.src('src/scss/*.scss')
        .pipe(sass())
        .pipe(prefix('last 2 versions'))
        .pipe(headerPipe())
        .pipe(gulp.dest('static/css'))
        .on('error', handleErrors);
});

// Compile

gulp.task('compile', ['build'], function() {
    return gulp.src(['static/vendor/**/*.js', '!**/*.min.js'])
        .pipe(uglify());

    // Do the concat, header, uglification
    // Rev the vendors
    // Rev the main app
    // CSS minification
});


// Git

var headerPipe;
gulp.task('git-rev', function(cb) {
    git.short(function(str) {
        headerPipe = function() {
            return header(headerFileContent, {pkg: pkg, gitRev: str, date: new Date()});
        };
        cb();
    });
});


// Defaults

gulp.task('default', ['watch']);

gulp.task('watch', ['build'], function() {
    gulp.watch(['src/index.html'], ['index'])
    gulp.watch(['src/scss/**/*.scss'], ['sass', 'index'])
    gulp.watch(['src/js/*.js'], ['app', 'index'])
    gulp.watch(['src/tpl/*.html'], ['template_cache', 'index'])
});

// Helpers

var loggerStart = function() {
    startTime = process.hrtime();
    gutil.log('Running', gutil.colors.green("'bundle'") + '...');
}

var loggerEnd = function() {
    var taskTime = process.hrtime(startTime);
    var prettyTime = prettyHrtime(taskTime);
    gutil.log('Finished', gutil.colors.green("'bundle'"), 'in', gutil.colors.magenta(prettyTime));
}

var handleErrors = function() {
    var args = Array.prototype.slice.call(arguments);

    // Send error to notification center with gulp-notify
    notify.onError({
	title: "Compile Error",
	message: "<%= error.message %>"
    }).apply(this, args);

    // Keep gulp from hanging on this task
    this.emit('end');
}
