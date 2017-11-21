/**
 * parseCSS3AnimationShorthand
 * Parses a CSS3 Animation statement into an object.
 * http://www.joelambert.co.uk
 *
 * Copyright 2011, Joe Lambert.
 * Free to use under the MIT license.
 * http://www.opensource.org/licenses/mit-license.php
 *
 * e.g. parseCSS3AnimationShorthand('boxRotate 600ms linear 2s');
 */
"use strict";
window.parseCSS3AnimationShorthand = function (statement) {
    var props = {
            name: null,
            duration: null,
            timingFunction: 'ease',
            delay: 0,
            iterationCount: 1,
            direction: 'normal'
        },
        remainder = statement,
        ms, t, fn, r, f, i, d;
    /* -- Get duration/delay -- */
    // Convert strings times in s or ms to ms integers
    ms = function (t) {
        return t.indexOf('ms') > -1 ? parseInt(t, 10) : parseInt(t, 10) * 1000;
    };
    t = statement.match(/[0-9]+m?s/g);
    if (t) {
        props.duration = t.length > 0 ? ms(t[0]) : props.duration;
        props.delay = t.length > 1 ? ms(t[1]) : props.delay;
        // Remove the found properties from the string
        remainder = remainder.replace(t[0], ''); // Replace the original found string not the cleansed one
    }
    /* -- Get timing function -- */
    fn = ['linear', 'ease', 'ease-in', 'ease-out', 'ease-in-out'];
    r = new RegExp('(' + fn.join('(\\s|$)|').replace('-', '\-') + '|cubic-bezier\\(.*?\\))');
    f = statement.match(r);
    if (f) {
        props.timingFunction = f.length > 0 ? f[0].replace(/\s+/g, '') : props.timingFunction;
        remainder = remainder.replace(props.timingFunction, '');
    }
    /* -- Get iteration count -- */
    i = statement.match(/(infinite|\s[0-9]+(\s|$))/);
    if (i) {
        props.iterationCount = i[0].replace(/\s+/g, '');
        remainder = remainder.replace(i[0], '');
    }
    /* -- Get direction -- */
    d = statement.match(/(normal|alternate)\s*$/);
    if (d) {
        props.direction = d[0].replace(/\s+/g, '');
        remainder = remainder.replace(d[0], '');
    }
    remainder = remainder.split(' ');
    if (remainder.length > 0) {
    	props.name = remainder[0];
    }
    return props;
};