Quicktable plugin for ckeditor 4.x
==========
###Version 1.0.5
- If the border gets disabled (`config.qtBorder = 0`) it isn't necessary to set the qtClass to `cke_show_border` anymore
	- Event `removeFormatCleanup` is fired if table gets inserted into the editor
- Small changes to the preview table design (no table border, cellspacing and cellpadding of 1)
- Adding option for preview table cell border
- Adding option for preview table cell background on hover
- Adding option for preview table cell size

###Version 1.0.4
- fix refactoring errors (table dimensions didn't get updated anymore)
	```javascript
	Cannot read property 'setText' of undefined
	```
	```javascript
	Cannot read property '$' of undefined 
	```

###Version 1.0.3
- short option values *(older ones are not supported anymore)*
- modify compiled sample so it refers to local ckeditor
- updated configuration sample
- example includes 2 editors (default and custom configuration)
- code refactoring to decrease code complexity and count of statements

###Version 1.0.2
- Adding cellspacing option
- Adding cellpadding option
- Change default width to table plugin default width

###Version 1.0.1
- change the sample to use ckeditor cdn
- adding a new class option to set the table attribute class
- change the style option type from string to object (for easier configuration)

###Version 1.0.0
- Initial Release
