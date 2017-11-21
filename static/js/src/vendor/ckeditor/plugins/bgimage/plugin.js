/**
 * Set background image for CKeditor
 * Author : Saeed Moqadam <phpro.ir@gmail.com>
 * Under MIT License
 */
CKEDITOR.plugins.add('bgimage',{
    lang:['en','fa','ru'],
    icons: 'bgimage',
    init:function (editor) {
        editor.addCommand('bgimage',new CKEDITOR.dialogCommand('bgImageDialog'));
        editor.ui.addButton(editor.lang.bgimage.bgImageTitle,{
            'label':editor.lang.bgimage.bgImageTitle,
            'command':'bgimage',
            'toolbar':'insert',
            icon: this.path + 'icons/bgimage.png'
        });
        CKEDITOR.dialog.add('bgImageDialog',this.path+'dialog/bgimageDialog.js');
    }
})
