CKEDITOR.dialog.add('bgImageDialog', function(editor) {
    return {
        title: editor.lang.bgimage.bgImageTitle,
        resizable: CKEDITOR.DIALOG_RESIZE_BOTH,
        minWidth: 500,
        minHeight: 200,
        onOk: function() {
            contents = editor.document.getBody().getHtml();
            matches = contents.match(/<div style="(.*)">((.|\n)*?)<\/div>/)
            // styled div already exists
            if(matches){
                contents = matches[2];
            }

            var dialog = this;
            var imageURL = dialog.getValueOf('tab1', 'imageURL');
            var repeat = dialog.getValueOf('tab1', 'repeat');
            var pos = dialog.getValueOf('tab1', 'pos')
            var blendMode = dialog.getValueOf('tab1', 'blend')
            var attachment = dialog.getValueOf('tab1', 'attachment')
            var width = dialog.getValueOf('tab1', 'width');
            var height = dialog.getValueOf('tab1', 'height');
            var div = '<div class="AA" style="';
            div += 'background-image:url(' + imageURL + ');';
            div += 'background-repeat:' + repeat + ';';
            div += 'background-position:' + pos + ';';
            div += 'background-blend-mode:' + blendMode + ';';
            div += 'background-attachment:' + attachment + ';';
            div += 'background-size:' + width +' '+height + ';';
            div += '">';
            div += contents;
            div += '</div>'
            editor.setData(div);
        },
        contents: [{
            id: 'tab1',
            label: editor.lang.bgimage.bgImageTitle,
            title: editor.lang.bgimage.bgImageTitle,
            accessKey: 'Q',
            elements: [{
                type: 'vbox',
                padding: 0,
                children: [{
                        type: 'hbox',
                        widths: ['280px', '100px;vertical-align: middle;'],
                        align: 'right',
                        styles :'',
                        children: [{
                            type: 'text',
                            label: editor.lang.bgimage.imageUrl,
                            id: 'imageURL',
                        }, {
                            type: 'button',
                            id: 'browse',
                            label: editor.lang.common.browseServer,
                            hidden: true,
                            filebrowser: 'tab1:imageURL'
                        }]
                    }]
            }, {
                type: 'vbox',
                padding: 0,
                children: [{
                        type: 'hbox',
                        widths: ['150px', '150px'],
                        align: 'right',
                        children: [{
                                type: 'select',
                                id: 'repeat',
                                label: editor.lang.bgimage.repeat,
                                items: [
                                    ['repeat'],
                                    ['no-repeat'],
                                    ['repeat-x'],
                                    ['repeat-y'],
                                ],
                                'default': 'repeat'
                            }, {
                                type: 'select',
                                id: 'attachment',
                                label: editor.lang.bgimage.attachment,
                                items: [
                                    ['scroll'],
                                    ['fixed'],
                                    ['local'],
                                ]
                            }]
                    }]
            }, {
                type: 'vbox',
                padding: 0,
                children: [{
                    type: 'hbox',
                    widths: ['150px', '150px'],
                    align: 'right',
                    children: [{
                        type: 'select',
                        id: 'blend',
                        label: editor.lang.bgimage.blendMode,
                        items: [
                            ['normal'],
                            ['multiply'],
                            ['screen'],
                            ['overlay'],
                            ['darken'],
                            ['lighten'],
                            ['color-dodge'],
                            ['saturation'],
                            ['color'],
                            ['luminosity'],
                        ],
                        style: 'float:left',
                        'default': 'normal'
                    }, {
                        type: 'select',
                        label: editor.lang.bgimage.position,
                        id: 'pos',
                        align: 'right',
                        items: [
                            ['left top'],
                            ['left center'],
                            ['left bottom'],
                            ['right top'],
                            ['right center'],
                            ['center top'],
                            ['center center'],
                            ['center center'],
                        ],
                        'default': 'left top'
                    }, ]
                }]
            },{
                            type: 'vbox',
                            padding: 0,
                            children: [{
                                    type: 'hbox',
                                    widths: ['150px', '150px'],
                                    align: 'right',
                                    children: [{
                                            type: 'select',
                                            id: 'repeat',
                                            label: editor.lang.bgimage.repeat,
                                            items: [
                                                ['repeat'],
                                                ['no-repeat'],
                                                ['repeat-x'],
                                                ['repeat-y'],
                                            ],
                                            'default': 'repeat'
                                        }, {
                                            type: 'select',
                                            id: 'attachment',
                                            label: editor.lang.bgimage.attachment,
                                            items: [
                                                ['scroll'],
                                                ['fixed'],
                                                ['local'],
                                            ]
                                        }]
                                }]
                        }, {
                            type: 'vbox',
                            padding: 0,
                            children: [{
                                type: 'hbox',
                                widths: ['150px', '150px'],
                                align: 'right',
                                children: [{
                                    type: 'text',
                                    id: 'width',
                                    label: editor.lang.bgimage.bgWidth,
                                    width:'50px',

                                }, {
                                    type: 'text',
                                    label: editor.lang.bgimage.bgHeight,
                                    id: 'height',
                                    align: 'right',
                                    width:'50px'
                                }]
                            }]
                        }]
        }],
    }
})
