#!/usr/bin/env python3
from reportlab.lib import colors
from reportlab.graphics.shapes import *
from reportlab.platypus import Table, TableStyle
import textwrap
from datetime import datetime
import os
from PIL import Image

from .PdfReport import PdfReport

# Adapted makeSEReport class for Gophish integration
class makeSEReport(PdfReport):

    def __init__(self):
        super().__init__()
        self.x_value = 20
        self.y_value = 700

        self.header = "Social Engineering Report"
        self.pageNumber = 0
        self.wrap_width = 90
        self.title_string = "Social Engineering Campaign Report"
        self.cName = ''
        self.testDate = ''
        self.urlString = ''
        self.emailList = []
        self.userClickRateChartList = []
        self.totalCredsCollected = 0
        self.userClickCount = 0
        self.totalClickCount = 0
        self.userWMultCredCount = 0
        self.figureCounter = 1
        self.blockedString = ''
        self.impersonated = ''
        self.domain = ''
        self.clientFolder = ''
        self.companyReportPath = ''

        # Default values for required attributes
        self.cA1 = ''
        self.cA2 = ''
        self.initialBlocked = 'n'
        self.testMethod = "externally using email links"
        self.navVecString = 'clicking on a link'
        self.credCSVFile = ''
        self.emailsMappedTxt = ''
        self.hitsReportTextFile = ''
        self.scanDataFile = ''
        self.screenshotFile = ''
        self.emailScreenshot = ''
        
        # Initialize data structures
        self.masterPWList = []
        self.knownPWList = []
        self.usersWCreds = 0
        self.credentialsDict = {}

    def setClientFolder(self, pathVar):
        self.clientFolder = pathVar

    def setDomain(self, domain):
        self.domain = domain

    def setClientName(self, name):
        self.clientName = name
        super().setClientName(name)

    def _process_mjset_data(self):
        """Process the MJSET-format data files"""
        # Process scan data
        if os.path.exists(self.scanDataFile):
            with open(self.scanDataFile, 'r') as sdf:
                for line in sdf:
                    if "Client Name = " in line:
                        self.cName = line[:-1].replace('Client Name = ', '')
                        super().setClientName(f"{self.cName}")
                    if "Scan Started = " in line:
                        self.testDate = line[:-1].replace('Scan Started = ', '')
                    if "Malicious URLs = " in line:
                        self.urlList = [line[:-1].replace('Malicious URLs = ', '')]
                        self.urlString = ', '.join(self.urlList)
                    if "Impersonated = " in line:
                        self.impersonated = line[:-1].replace('Impersonated = ', '')

        # Process email list
        self.emailList = []
        if os.path.exists(self.emailsMappedTxt):
            with open(self.emailsMappedTxt, 'r') as ef:
                for line in ef:
                    line = line.strip()
                    if line and '|' in line:
                        userEmail = line.split('|')[0]
                        if userEmail not in self.emailList:
                            self.emailList.append(userEmail)

        # Process credentials
        self.credentialsDict = {}
        self.masterPWList = []
        if os.path.exists(self.credCSVFile):
            with open(self.credCSVFile, "r") as cjf:
                for line in cjf:
                    if '|' in line:
                        listed = line.split('|')
                        if len(listed) >= 5:
                            user = listed[0]
                            userName = listed[1]
                            credential = listed[2]
                            sourceIP = listed[3]
                            timeStamp = listed[4].strip()
                            localList = [userName, credential, sourceIP, timeStamp]
                            
                            if credential:
                                self.masterPWList.append(credential)
                            
                            if user not in self.credentialsDict.keys():
                                self.credentialsDict[user] = [localList]
                            else:
                                temp_list = self.credentialsDict.get(user)
                                temp_list.append(localList)
                                self.credentialsDict[user] = temp_list

        # Process hits report
        self.userClickRateChartList = [("User/Email:", "Click/Visit Count:", "Credentials Count:")]
        self.userClickCount = 0
        self.totalClickCount = 0
        self.usersWCreds = 0
        self.totalCredsCollected = 0
        
        if os.path.exists(self.hitsReportTextFile):
            with open(self.hitsReportTextFile, 'r') as file:
                for line in file:
                    if line.strip() and '|' in line:
                        try:
                            parts = line.split('|')
                            if len(parts) >= 3:
                                user = parts[0]
                                hitcount = int(parts[1])
                                credentialsCount = int(parts[2])
                                
                                self.userClickRateChartList.append((user, str(hitcount), str(credentialsCount)))
                                
                                if hitcount > 0:
                                    self.userClickCount += 1
                                    self.totalClickCount += hitcount
                                
                                if credentialsCount > 0:
                                    self.usersWCreds += 1
                                    self.totalCredsCollected += credentialsCount
                        except (ValueError, IndexError):
                            continue

        # Calculate percentages
        if len(self.emailList) > 0:
            self.userClickPerc = round((self.userClickCount / len(self.emailList)) * 100, 1)
            self.userCredCountPercent = round((self.usersWCreds / len(self.emailList)) * 100, 1)
        else:
            self.userClickPerc = 0
            self.userCredCountPercent = 0

    def lineReplace(self, line):
        """Replace template variables in text lines"""
        replacements = {
            '<Client_Name>': self.cName,
            '<Date_of_Test>': str(self.testDate),
            '<Malicious_Domain>': str(self.domain),
            '<Email_List_Length>': str(len(self.emailList)),
            '<Credentials_Collected_Count>': str(self.totalCredsCollected),
            '<Employees_Who_Entered_Credentials_Count>': str(self.usersWCreds),
            '<was_or_was_not>': self.blockedString,
            '<Nav_Vector>': self.navVecString,
            '<Impersonated_Employee_Group>': self.impersonated,
            '<How_Test_Performed>': self.testMethod,
            '<ClientAddress1>': self.cA1,
            '<ClientAddress2>': self.cA2,
            '<userClickPercentage>': str(self.userClickPerc),
            '<credPercent>': str(self.userCredCountPercent),
            '<usercount>': str(len(self.emailList)),
            '<NumberOfPWs>': str(len(self.masterPWList)),
        }
        
        # Handle dynamic replacements
        if '<User_Click_Count>' in line:
            if self.userClickCount > 1:
                insertString = f"{self.userClickCount} unique employees"
            elif self.userClickCount == 1:
                insertString = f"only one ({self.userClickCount}) employee"
            else:
                insertString = f"any employee"
            line = line.replace('<User_Click_Count>', insertString)
        
        if '<Total_Click_Count>' in line:
            if self.totalClickCount > 1:
                insertString = f"were a total of {self.totalClickCount} visits"
            elif self.totalClickCount == 1:
                insertString = f"was only {self.totalClickCount} visit"
            else:
                insertString = f"were no visits"
            line = line.replace('<Total_Click_Count>', insertString)
        
        # Apply standard replacements
        for placeholder, value in replacements.items():
            line = line.replace(placeholder, value)
        
        return line

    def calculateImageRoom(self, imageFile):
        """Calculate image dimensions for PDF"""
        try:
            im = Image.open(imageFile)
            tempTuple = im.size  # (width,height) tuple
            tempWidth, tempHeight = tempTuple
            h_adj_mult = 320/tempHeight
            w_adj_mult = 575/tempWidth
            if h_adj_mult < w_adj_mult:
                tempWidth = tempWidth * h_adj_mult
                tempHeight = tempHeight * h_adj_mult
            else:
                tempWidth = tempWidth * w_adj_mult
                tempHeight = tempHeight * w_adj_mult
            tempX = 305 - tempWidth/2
            return(tempWidth, tempHeight, tempX)
        except:
            # Return default dimensions if image processing fails
            return (300, 200, 156)

    def insertTable(self, line):
        """Insert data tables into the report"""
        line = line.replace("#TABLE#", "")
        if "<Click_By_User_Details_Here>" in line:
            tableList = self.userClickRateChartList
            while len(tableList) > 32:
                tempList = tableList[:32]
                tempList2 = tableList[32:]
                tableList = [("User/Email:", "Click/Visit Count:", "Credentials Count:")]
                for item in tempList2:
                    tableList.append(item)
                self.newPage()
                self.printTable(tempList)
            if len(tableList) > 0:
                self.printTable(tableList)

    def newPage(self):
        """Create a new page with header"""
        self.c.showPage()
        self.writePageNumber()
        # Create the header
        self.c.setFillColorRGB(0.21, 0.56, 0.58)
        self.c.rect(0, 752, 612, 40, fill=True, stroke=False)
        # Write the title in the header
        super().f_pageHeader()
        self.c.drawString(self.x_value, 768, "Social Engineering Report")
        self.y_value = 700
        self.f_body()

    def insertImage(self, line):
        """Insert images into the report"""
        if "<Website_Image_Here>" in line:
            try:
                line = line.replace("<Website_Image_Here>", "")
                tempWidth, tempHeight, tempX = self.calculateImageRoom(self.screenshotFile)
                if self.y_value < tempHeight + 30:
                    self.newPage()
                self.y_value -= tempHeight
                self.c.drawImage(self.screenshotFile, tempX, self.y_value, width=tempWidth, height=tempHeight)
                self.y_value -= 14
            except:
                line = "<Website_Image_Here>"
                self.c.drawString(self.x_value + 150, self.y_value, line)
                self.y_value -= 14
        
        if "<Email_Image_Here>" in line:
            try:
                line = line.replace("<Email_Image_Here>", "")
                tempWidth, tempHeight, tempX = self.calculateImageRoom(self.emailScreenshot)
                if self.y_value < tempHeight + 30:
                    self.newPage()
                self.y_value -= tempHeight
                self.c.drawImage(self.emailScreenshot, tempX, self.y_value, width=tempWidth, height=tempHeight)
                self.y_value -= 14
            except:
                line = "<Email_Image_Here>"
                self.c.drawString(self.x_value + 150, self.y_value, line)
                self.y_value -= 14

    def generateLeaderPage(self):
        """Generate the main content pages"""
        self.newPage()
        
        # Create default template if it doesn't exist
        if not os.path.exists(self.leaderFile):
            self._create_default_template()
        
        with open(self.leaderFile, 'r') as file:
            for line in file:
                if line == '\n':
                    self.y_value -= 14
                    line = ''
                if self.y_value < 34:
                    self.newPage()
                if line.startswith("#IMAGE#"):
                    line = line.replace("#IMAGE#", "")
                    self.insertImage(line)
                    line = ''
                if line.startswith("#TABLE#"):
                    self.insertTable(line)
                    line = ''
                if line.startswith("Figure"):
                    super().f_bodySmall()
                    tempWidth = super().f_getTextWidth(line)
                    x_start = 305 - tempWidth/2
                    self.c.drawString(x_start, self.y_value, line)
                    self.f_body()
                    self.y_value -= 14
                    line = ''
                if line.startswith("#SUCCESS#"):
                    if len(self.masterPWList) > 0:
                        line = line.replace("#SUCCESS#", '')
                    else:
                        line = ''
                if line.startswith("#HEADER#"):
                    line = line.replace("#HEADER#", "")
                    super().f_bodyBold()
                    self.c.drawString(self.x_value, self.y_value, line)
                    self.f_body()
                    self.y_value -= 14
                    line = ''
                if line.startswith("#BLOCKED#"):
                    if self.initialBlocked == 'y':
                        line = line.replace("#BLOCKED#", '')
                    else:
                        line = ''
                if line != '':
                    self.f_body()
                    temp_string = self.lineReplace(line)
                    wrapped_text = textwrap.wrap(temp_string, width=100)
                    for lines in wrapped_text:
                        if self.y_value < 60:
                            self.newPage()
                        self.c.drawString(self.x_value, self.y_value, lines)
                        self.y_value -= 14
        
        # Add signature section
        self.f_header()
        self.y_value -= 56
        try:
            self.c.drawImage(self.mjsignature, self.x_value, self.y_value, width=200, height=50)
        except:
            self.c.drawString(self.x_value, self.y_value, "Mauldin & Jenkins")
        self.y_value -= 28
        self.f_body()
        self.c.drawString(self.x_value, self.y_value, 'Chattanooga, TN')
        self.y_value -= 14
        now = f"{datetime.now():%B %d, %Y}"
        self.c.drawString(self.x_value, self.y_value, '{0:}'.format(now))

    def _create_default_template(self):
        """Create a default report template"""
        template_dir = os.path.dirname(self.leaderFile)
        os.makedirs(template_dir, exist_ok=True)
        
        with open(self.leaderFile, 'w') as f:
            f.write("To the Management of <Client_Name>:\n\n")
            f.write("#HEADER#Purpose:\n\n")
            f.write("The purpose of Social Engineering Testing is to evaluate the extent to which employees comply with the security policies and protocols set by management. Through testing, an organization can gather insights into the potential susceptibility of its employees to unauthorized access attempts, breach of security protocols, or disclosure of sensitive information.\n\n")
            
            f.write("#HEADER#Scope:\n\n")
            f.write("Mauldin & Jenkins, LLC. performed a Social Engineering Phishing Campaign against <Client_Name>'s network and employees on <Date_of_Test>. The process was designed to initially test the technical controls in place to block this attack vector, then focus on the employees and their adherence to policies, procedures and training. The test was performed <How_Test_Performed>.\n\n")
            
            f.write("#HEADER#Procedures:\n\n")
            f.write("Using an internally developed, custom social engineering tool, the auditor created an email phishing campaign designed to lure <Client_Name>'s employees into navigating to our 'malicious' site. <Client_Name> provided us with a list of <Email_List_Length> employees and their email addresses.\n\n")
            f.write("We posed as <Impersonated_Employee_Group>, and directed the employees to visit our site by <Nav_Vector>. The auditor sent communications to <Email_List_Length> of <Client_Name>'s employees.\n\n")
            
            f.write("#IMAGE#<Email_Image_Here>\n")
            f.write("Figure 1 – Phishing Email Example\n\n")
            
            f.write("The auditor configured the following website <Malicious_Domain> containing a sign-in prompt to support the pre-determined attack vector, as shown in Figure 2. The auditor's communications <was_or_was_not> initially blocked by <Client_Name>'s existing network defenses.\n\n")
            
            f.write("#IMAGE#<Website_Image_Here>\n")
            f.write("Figure 2 – Phishing Website Screenshot\n\n")
            
            f.write("#HEADER#Results Overview\n\n")
            f.write("• Total employees targeted: <Email_List_Length>\n")
            f.write("• Employees who clicked the link: <User_Click_Count> (<userClickPercentage>%)\n")
            f.write("• Employees who submitted credentials: <Employees_Who_Entered_Credentials_Count> (<credPercent>%)\n")
            f.write("• Total credentials collected: <Credentials_Collected_Count>\n\n")
            
            f.write("#SUCCESS##HEADER#Password Analysis\n\n")
            f.write("#SUCCESS#A total of <NumberOfPWs> passwords were collected during this assessment.\n\n")
            
            f.write("#HEADER#Detailed Results\n\n")
            f.write("The following table shows the detailed click and credential submission data:\n\n")
            f.write("#TABLE#<Click_By_User_Details_Here>\n\n")

    def printTable(self, passed_list):
        """Print a data table in the report"""
        table = Table(passed_list)
        table.setStyle(TableStyle([
            ("BACKGROUND", (0, 0), (-1, 0), colors.HexColor("#379095")),
            ("TEXTCOLOR", (0, 0), (-1, 0), colors.whitesmoke),
            ("ALIGN", (0, 0), (-1, -1), "CENTER"),
            ("FONTNAME", (0, 0), (-1, 0), "Helvetica-Bold"),
            ("FONTSIZE", (0, 0), (-1, 0), 12),
            ("BOTTOMPADDING", (0, 0), (-1, 0), 12),
            ("BACKGROUND", (0, 1), (-1, -1), colors.HexColor("#e0e0e0")),
            ("TEXTCOLOR", (0, 1), (-1, -1), colors.black),
            ("FONTNAME", (0, 1), (-1, -1), "Helvetica"),
            ("FONTSIZE", (0, 1), (-1, -1), 10),
            ("ALIGN", (0, 0), (-1, -1), "CENTER"),
            ('GRID', (0, 0), (-1, -1), 1, colors.HexColor("#379095")),
        ]))
        table.wrapOn(self.c, 200, 200)
        table_height = table._height
        table_width = table._width
        if self.y_value - table_height < 48:
            self.newPage()
        table.drawOn(self.c, 305 - table_width/2, self.y_value - table_height)
        self.y_value = self.y_value - table_height - 14

    def writeReport(self):
        """Main report writing function"""
        # Set up output file
        super().setOutputFile(f"{self.companyReportPath}SocialEngExecReport.pdf")
        
        # Process the data
        self._process_mjset_data()
        
        # Set up leader file
        super().setLeaderFile('SE_Leader.txt')
        
        # Write the report
        super().writeCoverPage()
        self.generateLeaderPage()
        super().saveReport()
        
        print(f"Executive report generated: {self.companyReportPath}SocialEngExecReport.pdf")
