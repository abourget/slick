var gulp = require('gulp');

//var browserSync  = require('browser-sync');
//var browserify   = require('browserify');
//var webpack      = require('webpack');
//var watchify     = require('watchify');
var source       = require('vinyl-source-stream');
var changed      = require('gulp-changed');
var imagemin     = require('gulp-imagemin');
var compass      = require('gulp-compass');
var notify       = require('gulp-notify');
var gutil        = require('gulp-util');
var prettyHrtime = require('pretty-hrtime');
var startTime;


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

gulp.task('images', function() {
    var dest = './build/images';

    return gulp.src('./src/images/**')
        .pipe(changed(dest)) // Ignore unchanged files
        .pipe(imagemin()) // Optimize
        .pipe(gulp.dest(dest));
});

gulp.task('browserSync', ['build'], function() {
    browserSync.init(['build/**'], {
        server: {
            baseDir: 'build'
        }
    });
});

gulp.task("webpack", function() {
    return gulp.src('src/entry.js')
        .pipe(webpack({ /* webpack configuration */ }))
        .pipe(gulp.dest('dist/'));
});

gulp.task('browserify', function() {

    var bundleMethod = global.isWatching ? watchify : browserify;

    var bundler = bundleMethod({
        // Specify the entry point of your app
        entries: ['./src/javascript/app.coffee'],
        // Add file extentions to make optional in your requires
        extensions: ['.coffee', '.hbs'],
        // Enable source maps!
        debug: true
    });

    var bundle = function() {
        // Log when bundling starts
        loggerStart();

        return bundler
            .bundle()
        // Report compile errors
            .on('error', handleErrors)
        // Use vinyl-source-stream to make the
        // stream gulp compatible. Specifiy the
        // desired output filename here.
            .pipe(source('app.js'))
        // Specify the output destination
            .pipe(gulp.dest('./build/'))
        // Log when bundling completes!
            .on('end', loggerEnd);
    };

    if(global.isWatching) {
        // Rebundle with watchify on changes.
        bundler.on('update', bundle);
    }

    return bundle();
});


gulp.task('setWatch', function() {
    global.isWatching = true;
});

gulp.task('watch', ['setWatch', 'browserSync'], function() {
    gulp.watch('src/sass/**', ['compass']);
    gulp.watch('src/images/**', ['images']);
    gulp.watch('src/htdocs/**', ['copy']);
    // Note: The browserify task handles js recompiling with watchify
});


gulp.task('build', ['browserify', 'compass', 'images', 'copy']);

gulp.task('copy', function() {
    return gulp.src('src/htdocs/**')
        .pipe(gulp.dest('build'));
});

gulp.task('default', ['watch']);

gulp.task('compass', function() {
    return gulp.src('./src/sass/*.sass')
        .pipe(compass({
            project: path.join(__dirnam, 'assets'),
            css: 'build',
            sass: 'src/sass'
        }))
        .pipe(prefix('last 2 versions'))
        .pipe(gulp.dest('app/css'))
        .on('error', handleErrors);
});

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
