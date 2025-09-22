#!/usr/bin/env python3
from reportlab.pdfgen import canvas
from reportlab.lib.pagesizes import letter
from reportlab.lib import colors
from reportlab.graphics.shapes import *
from reportlab.pdfbase import pdfmetrics
from reportlab.pdfbase.ttfonts import TTFont
import textwrap, datetime
import os

'''
Adapted from MJSET PdfReport for Gophish integration
Original by Pierce Loechl for Mauldin and Jenkins IT Advisory.
'''

class PdfReport:
    def __init__(self):
        
        # initialize class variables
        self.csvFile = ""
        self.outputFile = "output.pdf"
        self.c = None
        self.pageNumber = 1
        self.directory = os.path.join(os.path.dirname(__file__), '..', 'assets') + '/'
        self.clientName = ''
        self.date1 = ''
        self.date2 = ''
        self.reportName = ''
        self.remediated_count = ''
        self.not_remediated_count = ''
        self.new_piid_count = ''
        self.criticality = ''

        # set up some image variables - use placeholder paths for now
        self.mjCover = "{0:}images/mjcover.png".format(self.directory)
        self.mjLogo = "{0:}images/mjlogo.png".format(self.directory)
        self.mjLogoSmall = "{0:}images/mjlogosmall.png".format(self.directory)
        self.mjsignature = "{0:}images/mjsignature.png".format(self.directory)
        self.greenCheck = '{0:}images/checkmark-48.png'.format(self.directory)
        self.blueCircle = "{0:}images/blueCircle-48.png".format(self.directory)

        # set up the fonts - use system fonts as fallback
        try:
            pdfmetrics.registerFont(TTFont('Playfair', '{0:}fonts/PlayfairDisplay-Regular.ttf'.format(self.directory)))
            pdfmetrics.registerFont(TTFont('PlayfairBD', '{0:}fonts/PlayfairDisplay-Bold.ttf'.format(self.directory)))
            pdfmetrics.registerFont(TTFont('Forum', '{0:}fonts/Forum-Regular.ttf'.format(self.directory)))
            pdfmetrics.registerFont(TTFont('Roboto', '{0:}fonts/Roboto-Regular.ttf'.format(self.directory)))
            pdfmetrics.registerFont(TTFont('RobotoBd', '{0:}fonts/Roboto-Bold.ttf'.format(self.directory)))
        except:
            # Fallback to system fonts if custom fonts are not available
            pass

        # set up the colors
        self.mjBlue = colors.Color(red=(55/255), green=(144/255), blue=(149/255))
        self.mjDarkBlue = colors.Color(red=(47/255), green=(61/255), blue=(76/255))
        self.mjTeal = colors.Color(red=(0/255), green=(240/255), blue=(212/255))
        self.mjGray = colors.Color(red=(79/255), green=(81/255), blue=(82/255))
        self.mjGold = colors.Color(red=(200/255), green=(178/255), blue=(115/255))
        self.orange = colors.Color(red=(255/255), green=(192/255), blue=(0/255))

    # getters and setters
        
    def getBlueCircle(self):
        self.blueCircle = self.blueCircle
        
    def getGreenCheck(self):
        self.greenCheck = self.greenCheck

    def setCSVFile(self, csvFile):
        self.csvFile = csvFile

    def setOutputFile(self, outputFile):
        self.outputFile = outputFile
        self.createCanvas()

    def setClientName(self, clientName):
        self.clientName = clientName

    def setReportName(self, reportName):
        self.reportName = reportName

    def setDate1(self, date1):
        self.date1 = str(date1)

    def setDate2(self, date2):
        self.date2 = str(date2)

    def setLeaderFile(self, leaderFile):
        self.leaderFile = os.path.join(os.path.dirname(__file__), '..', 'templates', leaderFile)

    def getCSVFile(self):
        return self.csvFile
    
    def getOutputFile(self):
        return self.outputFile
    
    def getCanvas(self):
        return self.c

    def setRemediatedList(self, remediated_count):
        self.remediated_count = remediated_count

    def setNotRemediatedList(self, not_remediated_count):
        self.not_remediated_count = not_remediated_count

    def setNewList(self, new_piid_count):
        self.new_piid_count = new_piid_count

    def setCriticality(self, criticality):
        self.criticality = criticality
    
    '''
            FONT METHODS
    -------------------------------------
    These methods set up standard fonts and font sizes for use accross all types of PDF reports.
    This is to ensure brand continuity and to make it easier to change fonts and sizes in the future.
    If you need to change font size or color, you can set the optional variables. 
    Of course, you can also set the font and size manually if you need to.
    -------------------------------------
    '''
    
    def f_title(self, color=colors.white, size=25):
        self.c.setFillColor(color)
        try:
            self.c.setFont("Playfair", size)
        except:
            self.c.setFont("Helvetica-Bold", size)

    def f_pageHeader(self, color=colors.white, size=14):
        self.c.setFillColor(color)
        try:
            self.c.setFont("PlayfairBD", size)
        except:
            self.c.setFont("Helvetica-Bold", size)

    def f_header(self, color=colors.black, size=20):
        self.c.setFillColor(color)
        try:
            self.c.setFont("RobotoBd", size)
        except:
            self.c.setFont("Helvetica-Bold", size)

    def f_body(self, color=colors.black, size=12):
        self.c.setFillColor(color)
        try:
            self.c.setFont("Roboto", size)
        except:
            self.c.setFont("Helvetica", size)

    def f_bodyBold(self, color=colors.black, size=12):
        self.c.setFillColor(color)
        try:
            self.c.setFont("RobotoBd", size)
        except:
            self.c.setFont("Helvetica-Bold", size)

    def f_bodySmall(self, color=colors.black, size=10):
        self.c.setFillColor(color)
        try:
            self.c.setFont("Roboto", size)
        except:
            self.c.setFont("Helvetica", size)

    def f_bodySmallBold(self, color=colors.black, size=10):
        self.c.setFillColor(color)
        try:
            self.c.setFont("RobotoBd", size)
        except:
            self.c.setFont("Helvetica-Bold", size)

    def f_getTextWidth(self, text):
        # this is a handy function for getting the text width of the currently set font and size when using the f_ methods.
        currentFont = self.c._fontname
        currentSize = self.c._fontsize
        return self.c.stringWidth(text, currentFont, currentSize)

    '''
            MAIN METHODS
    -------------------------------------
    These methods are the main methods that are called to create the PDF report.

    -------------------------------------
    '''

    def createCanvas(self):
        # Create our canvas object that the PDF will be written to.
        # this method must be called after the outputFile is set, or it will be given a default canvas name.
        # setting the outputFile will call this method automatically.
        self.c = canvas.Canvas(self.outputFile, pagesize=letter)

    def writeCoverPage(self):
        # Try to draw cover image, fallback to text if not available
        try:
            self.c.drawImage(self.mjCover, 0, 0, width=612, height=792)
        except:
            # Fallback: create a simple cover page
            self.c.setFillColor(self.mjBlue)
            self.c.rect(0, 0, 612, 792, fill=True, stroke=False)
        
        self.f_title()
        text_width = self.f_getTextWidth(self.title_string)
        self.c.drawString((612 - text_width) / 2, 550, self.title_string)

    def writePageNumber(self):
        if self.pageNumber == '':
            self.pageNumber = 0
        else:
            self.pageNumber += 1
        # Write the page number to the bottom right of the page.
        self.c.setFont("Helvetica", 10)
        self.c.drawString(593, 10, str(self.pageNumber))
        # Try to draw the M&J logo next to it, fallback to text
        try:
            self.c.drawImage(self.mjLogoSmall, 558, 0, width=30, height=30)
        except:
            self.c.drawString(520, 10, "M&J")
        
    def generateLeaderPage(self):
        self.c.showPage()
        self.x_value = 30
        self.y_value = 700
        self.pageNumber = ''
        self.writePageNumber()
        self.f_body()
        
        # Check if leader file exists, create default content if not
        if not os.path.exists(self.leaderFile):
            self._create_default_leader_content()
        
        with open(self.leaderFile, 'r') as file:
            for line in file:
                if line == '\n':
                    self.y_value -= 14
                elif line != '':
                    temp_string = line.replace('clientName', self.clientName).replace('scanDate1', self.date1).replace('scanDate2', self.date2)
                    temp_string = temp_string.replace('reportName', self.reportName).replace('remediatedPiidCount', str(self.remediated_count))
                    temp_string = temp_string.replace('notRemediatedPiidCount', str(self.not_remediated_count)).replace('newVulnerabilities', str(self.new_piid_count))
                    temp_string = temp_string.replace('reportCriticality', str(self.criticality))
                    wrapped_text = textwrap.wrap(temp_string, width=100)
                    for lines in wrapped_text:
                        self.c.drawString(self.x_value, self.y_value, lines)
                        self.y_value -= 14
        self.f_header()
        self.y_value -= 56
        
        # Try to draw signature, fallback to text
        try:
            self.c.drawImage(self.mjsignature, self.x_value, self.y_value, width=200, height=50)
        except:
            self.c.drawString(self.x_value, self.y_value, "Mauldin & Jenkins")
        
        self.y_value -= 28
        self.f_body()
        self.c.drawString(self.x_value, self.y_value, 'Chattanooga, TN')
        self.y_value -= 14
        now = f"{datetime.datetime.now():%B %d, %Y}"
        self.c.drawString(self.x_value, self.y_value, '{0:}'.format(now))

    def _create_default_leader_content(self):
        """Create default leader page content if template doesn't exist"""
        os.makedirs(os.path.dirname(self.leaderFile), exist_ok=True)
        with open(self.leaderFile, 'w') as f:
            f.write("Social Engineering Assessment Report\n\n")
            f.write("Client: clientName\n")
            f.write("Assessment Date: scanDate1\n\n")
            f.write("This report contains the results of the social engineering assessment conducted for clientName.\n\n")
            f.write("The assessment was designed to evaluate the organization's susceptibility to social engineering attacks.\n\n")

    def saveReport(self):
        # Save the report to the file. This method is called when the report has been built and is ready to be saved.
        self.c.save()
