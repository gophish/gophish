var TEMPLATE_TAGS = [{
        id: 1,
        name: 'RId',
        description: 'The unique ID for the recipient.'
    },
    {
        id: 2,
        name: 'FirstName',
        description: 'The recipient\'s first name.'
    },
    {
        id: 3,
        name: 'LastName',
        description: 'The recipient\'s last name.'
    },
    {
        id: 4,
        name: 'Position',
        description: 'The recipient\'s position.'
    },
    {
        id: 5,
        name: 'From',
        description: 'The address emails are sent from.'
    },
    {
        id: 6,
        name: 'TrackingURL',
        description: 'The URL to track emails being opened.'
    },
    {
        id: 7,
        name: 'Tracker',
        description: 'An HTML tag that adds a hidden tracking image (recommended instead of TrackingURL).'
    },
    {
        id: 8,
        name: 'URL',
        description: 'The URL to your Gophish listener.'
    },
    {
        id: 9,
        name: 'BaseURL',
        description: 'The base URL with the path and rid parameter stripped. Useful for making links to static files.'
    }
];

var textTestCallback = function (range) {
    if (!range.collapsed) {
        return null;
    }

    return CKEDITOR.plugins.textMatch.match(range, matchCallback);
}

var matchCallback = function (text, offset) {
    var pattern = /\{{2}\.?([A-z]|\})*$/,
        match = text.slice(0, offset)
        .match(pattern);

    if (!match) {
        return null;
    }

    return {
        start: match.index,
        end: offset
    };
}

/**
 * 
 * @param {regex} matchInfo - The matched text object
 * @param {function} callback - The callback to execute with the matched data
 */
var dataCallback = function (matchInfo, callback) {
    var data = TEMPLATE_TAGS.filter(function (item) {
        var itemName = '{{.' + item.name.toLowerCase() + '}}';
        return itemName.indexOf(matchInfo.query.toLowerCase()) == 0;
    });

    callback(data);
}

/**
 * 
 * @param {CKEditor} editor - The CKEditor instance.
 * 
 * Installs the autocomplete plugin to the CKEditor.
 */
var setupAutocomplete = function (editor) {
    editor.on('instanceReady', function (evt) {
        var itemTemplate = '<li data-id="{id}">' +
            '<div><strong class="item-title">{name}</strong></div>' +
            '<div><i>{description}</i></div>' +
            '</li>',
            outputTemplate = '[[.{name}]]';

        var autocomplete = new CKEDITOR.plugins.autocomplete(evt.editor, {
            textTestCallback: textTestCallback,
            dataCallback: dataCallback,
            itemTemplate: itemTemplate,
            outputTemplate: outputTemplate
        });

        // We have to use brackets for the output template tag and 
        // then manually replace them due to the way CKEditor's 
        // templating works.
        autocomplete.getHtmlToInsert = function (item) {
            var parsedTemplate = this.outputTemplate.output(item);
            parsedTemplate = parsedTemplate.replace("[[", "{{").replace("]]", "}}")
            return parsedTemplate
        }
    });
}