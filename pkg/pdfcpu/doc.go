/*

Package pdfcpu is a simple PDF processing library written in Go supporting encryption.
It provides an API and a command line interface. Supported are all versions up to PDF 1.7 (ISO-32000).

The commands are:

	validate	validate PDF against PDF 32000-1:2008 (PDF 1.7)
	optimize	optimize PDF by getting rid of redundant page resources
	split		split multi-page PDF into several PDFs according to split span
	merge		concatenate 2 or more PDFs
	extract		extract images, fonts, content, pages or metadata
	trim		create trimmed version with selected pages
	stamp		add text or image stamp to selected pages
	watermark	add text or image watermark for selected pages
	import		import/convert images
	nup			rearrange pages/images for reduced number of pages
	grid		rearrange pages/images for enhanced browsing experience
	rotate		rotate pages
	attachments	list, add, remove, extract embedded file attachments
	permissions	list, add user access permissions
	encrypt		set password protection
	decrypt		remove password protection
	changeupw	change user password
	changeopw	change owner password
	paper		print list of supported paper sizes
	version		print version

*/
package pdfcpu