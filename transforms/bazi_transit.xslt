<?xml version="1.0" encoding="UTF-8"?>
<!--
  BaZi Transit XSLT

  Consumes TransitChart XML. Computes:
  - Transit Four Pillars (year/month/day/hour from Moment)
  - Natal Four Pillars (from Natal/BaseChart/Time)
  - Transit pillar × natal pillar stem/branch interactions
  - Ten God analysis of transit Day Master vs natal Day Master
  - Luck pillar context (which luck pillar the transit falls in)

  Does NOT use planetary longitudes — purely stem/branch math.
-->
<xsl:stylesheet xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
                xmlns:math="http://exslt.org/math"
                xmlns:bazi="urn:empirical:bazi"
                version="1.0">

  <!-- ── Heavenly Stems ───────────────────────────────────────────────── -->
  <xsl:variable name="stems" select="'Jia,Yi,Bing,Ding,Wu,Ji,Geng,Xin,Ren,Gui'"/>
  <xsl:variable name="stemElements" select="'Wood,Wood,Fire,Fire,Earth,Earth,Metal,Metal,Water,Water'"/>
  <xsl:variable name="stemYinYang" select="'Yang,Yin,Yang,Yin,Yang,Yin,Yang,Yin,Yang,Yin'"/>

  <!-- ── Earthly Branches ──────────────────────────────────────────────── -->
  <xsl:variable name="branches" select="'Zi,Chou,Yin,Mao,Chen,Si,Wu,Wei,Shen,You,Xu,Hai'"/>
  <xsl:variable name="branchElements" select="'Water,Earth,Wood,Wood,Earth,Fire,Fire,Earth,Metal,Metal,Earth,Water'"/>
  <xsl:variable name="branchAnimals" select="'Rat,Ox,Tiger,Rabbit,Dragon,Snake,Horse,Goat,Monkey,Rooster,Dog,Pig'"/>

  <!-- ── Hidden Stems (per branch index) ───────────────────────────────── -->
  <xsl:variable name="hiddenStems0" select="'9'"/>       <!-- Zi → Gui -->
  <xsl:variable name="hiddenStems1" select="'5,9,7'"/>   <!-- Chou → Ji, Gui, Xin -->
  <xsl:variable name="hiddenStems2" select="'0,2,4'"/>   <!-- Yin → Jia, Bing, Wu -->
  <xsl:variable name="hiddenStems3" select="'1'"/>       <!-- Mao → Yi -->
  <xsl:variable name="hiddenStems4" select="'4,1,9'"/>   <!-- Chen → Wu, Yi, Gui -->
  <xsl:variable name="hiddenStems5" select="'2,6,4'"/>   <!-- Si → Bing, Geng, Wu -->
  <xsl:variable name="hiddenStems6" select="'3,5'"/>     <!-- Wu → Ding, Ji -->
  <xsl:variable name="hiddenStems7" select="'5,3,1'"/>   <!-- Wei → Ji, Ding, Yi -->
  <xsl:variable name="hiddenStems8" select="'6,8,4'"/>   <!-- Shen → Geng, Ren, Wu -->
  <xsl:variable name="hiddenStems9" select="'7'"/>       <!-- You → Xin -->
  <xsl:variable name="hiddenStems10" select="'4,7,3'"/>  <!-- Xu → Wu, Xin, Ding -->
  <xsl:variable name="hiddenStems11" select="'8,0'"/>    <!-- Hai → Ren, Jia -->

  <!-- ── Ten Gods names ───────────────────────────────────────────────── -->
  <xsl:variable name="tenGodsSame" select="'Bi Jian,Jie Cai'"/>
  <xsl:variable name="tenGodsProduce" select="'Zheng Yin,Pian Yin'"/>
  <xsl:variable name="tenGodsProduced" select="'Shi Shen,Shang Guan'"/>
  <xsl:variable name="tenGodsControl" select="'Qi Sha,Zheng Guan'"/>
  <xsl:variable name="tenGodsControlled" select="'Pian Cai,Zheng Cai'"/>

  <!-- ══════════════════════════════════════════════════════════════════════
       ROOT TEMPLATE
       ════════════════════════════════════════════════════════════════════ -->
  <xsl:template match="/TransitChart">
    <bazi:TransitReport xmlns:bazi="urn:empirical:bazi">
      <bazi:Name><xsl:value-of select="Identity/Name"/></bazi:Name>
      <bazi:TransitDate>
        <xsl:value-of select="Moment/Year"/>-<xsl:value-of select="format-number(Moment/Month,'00')"/>-<xsl:value-of select="format-number(Moment/Day,'00')"/>
      </bazi:TransitDate>

      <!-- ══════════════════════════════════════════════════════════════════
           COMPUTE NATAL FOUR PILLARS
           ════════════════════════════════════════════════════════════════ -->
      <xsl:variable name="natalYear" select="Natal/Time/Year"/>
      <xsl:variable name="natalMonth" select="Natal/Time/Month"/>
      <xsl:variable name="natalDay" select="Natal/Time/Day"/>
      <xsl:variable name="natalHour" select="Natal/Time/Hour"/>
      <xsl:variable name="natalDayJD" select="Natal/Time/DayJD"/>

      <!-- Natal Year Pillar -->
      <xsl:variable name="natalYearStem" select="($natalYear - 4) mod 10"/>
      <xsl:variable name="natalYearBranch" select="($natalYear - 4) mod 12"/>

      <!-- Natal Month Pillar (starts at Yin=2 for Jia year) -->
      <xsl:variable name="natalMonthStem" select="(($natalYearStem mod 5) * 2 + $natalMonth) mod 10"/>
      <xsl:variable name="natalMonthBranch" select="($natalMonth + 1) mod 12"/>

      <!-- Natal Day Pillar (from DayJD) -->
      <xsl:variable name="natalDayStem" select="($natalDayJD + 5) mod 10"/>
      <xsl:variable name="natalDayBranch" select="($natalDayJD + 3) mod 12"/>

      <!-- Natal Hour Pillar -->
      <xsl:variable name="natalHourBranch" select="floor(($natalHour + 1) div 2) mod 12"/>
      <xsl:variable name="natalHourStem" select="(($natalDayStem mod 5) * 2 + $natalHourBranch) mod 10"/>

      <!-- Natal Day Master -->
      <xsl:variable name="natalDMStem" select="$natalDayStem"/>
      <xsl:variable name="natalDMElement">
        <xsl:call-template name="nthToken">
          <xsl:with-param name="list" select="$stemElements"/>
          <xsl:with-param name="n" select="$natalDMStem + 1"/>
        </xsl:call-template>
      </xsl:variable>
      <xsl:variable name="natalDMYinYang">
        <xsl:call-template name="nthToken">
          <xsl:with-param name="list" select="$stemYinYang"/>
          <xsl:with-param name="n" select="$natalDMStem + 1"/>
        </xsl:call-template>
      </xsl:variable>

      <!-- ══════════════════════════════════════════════════════════════════
           COMPUTE TRANSIT FOUR PILLARS
           ════════════════════════════════════════════════════════════════ -->
      <xsl:variable name="transitYear" select="Moment/Year"/>
      <xsl:variable name="transitMonth" select="Moment/Month"/>
      <xsl:variable name="transitDay" select="Moment/Day"/>
      <xsl:variable name="transitHour" select="Moment/Hour"/>
      <xsl:variable name="transitDayJD" select="Moment/DayJD"/>

      <!-- Transit Year Pillar -->
      <xsl:variable name="transitYearStem" select="($transitYear - 4) mod 10"/>
      <xsl:variable name="transitYearBranch" select="($transitYear - 4) mod 12"/>

      <!-- Transit Month Pillar -->
      <xsl:variable name="transitMonthStem" select="(($transitYearStem mod 5) * 2 + $transitMonth) mod 10"/>
      <xsl:variable name="transitMonthBranch" select="($transitMonth + 1) mod 12"/>

      <!-- Transit Day Pillar -->
      <xsl:variable name="transitDayStem" select="($transitDayJD + 5) mod 10"/>
      <xsl:variable name="transitDayBranch" select="($transitDayJD + 3) mod 12"/>

      <!-- Transit Hour Pillar -->
      <xsl:variable name="transitHourBranch" select="floor(($transitHour + 1) div 2) mod 12"/>
      <xsl:variable name="transitHourStem" select="(($transitDayStem mod 5) * 2 + $transitHourBranch) mod 10"/>

      <!-- Transit Day Master -->
      <xsl:variable name="transitDMStem" select="$transitDayStem"/>
      <xsl:variable name="transitDMElement">
        <xsl:call-template name="nthToken">
          <xsl:with-param name="list" select="$stemElements"/>
          <xsl:with-param name="n" select="$transitDMStem + 1"/>
        </xsl:call-template>
      </xsl:variable>

      <!-- ══════════════════════════════════════════════════════════════════
           NATAL PILLARS
           ════════════════════════════════════════════════════════════════ -->
      <bazi:NatalPillars>
        <bazi:YearPillar>
          <bazi:Stem>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$stems"/>
              <xsl:with-param name="n" select="$natalYearStem + 1"/>
            </xsl:call-template>
          </bazi:Stem>
          <bazi:Branch>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$branches"/>
              <xsl:with-param name="n" select="$natalYearBranch + 1"/>
            </xsl:call-template>
          </bazi:Branch>
        </bazi:YearPillar>
        <bazi:MonthPillar>
          <bazi:Stem>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$stems"/>
              <xsl:with-param name="n" select="$natalMonthStem + 1"/>
            </xsl:call-template>
          </bazi:Stem>
          <bazi:Branch>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$branches"/>
              <xsl:with-param name="n" select="$natalMonthBranch + 1"/>
            </xsl:call-template>
          </bazi:Branch>
        </bazi:MonthPillar>
        <bazi:DayPillar>
          <bazi:Stem>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$stems"/>
              <xsl:with-param name="n" select="$natalDayStem + 1"/>
            </xsl:call-template>
          </bazi:Stem>
          <bazi:Branch>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$branches"/>
              <xsl:with-param name="n" select="$natalDayBranch + 1"/>
            </xsl:call-template>
          </bazi:Branch>
        </bazi:DayPillar>
        <bazi:HourPillar>
          <bazi:Stem>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$stems"/>
              <xsl:with-param name="n" select="$natalHourStem + 1"/>
            </xsl:call-template>
          </bazi:Stem>
          <bazi:Branch>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$branches"/>
              <xsl:with-param name="n" select="$natalHourBranch + 1"/>
            </xsl:call-template>
          </bazi:Branch>
        </bazi:HourPillar>
        <bazi:DayMaster>
          <bazi:Stem>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$stems"/>
              <xsl:with-param name="n" select="$natalDMStem + 1"/>
            </xsl:call-template>
          </bazi:Stem>
          <bazi:Element><xsl:value-of select="$natalDMElement"/></bazi:Element>
          <bazi:YinYang><xsl:value-of select="$natalDMYinYang"/></bazi:YinYang>
        </bazi:DayMaster>
      </bazi:NatalPillars>

      <!-- ══════════════════════════════════════════════════════════════════
           TRANSIT PILLARS
           ════════════════════════════════════════════════════════════════ -->
      <bazi:TransitPillars>
        <bazi:YearPillar>
          <bazi:Stem>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$stems"/>
              <xsl:with-param name="n" select="$transitYearStem + 1"/>
            </xsl:call-template>
          </bazi:Stem>
          <bazi:Branch>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$branches"/>
              <xsl:with-param name="n" select="$transitYearBranch + 1"/>
            </xsl:call-template>
          </bazi:Branch>
        </bazi:YearPillar>
        <bazi:MonthPillar>
          <bazi:Stem>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$stems"/>
              <xsl:with-param name="n" select="$transitMonthStem + 1"/>
            </xsl:call-template>
          </bazi:Stem>
          <bazi:Branch>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$branches"/>
              <xsl:with-param name="n" select="$transitMonthBranch + 1"/>
            </xsl:call-template>
          </bazi:Branch>
        </bazi:MonthPillar>
        <bazi:DayPillar>
          <bazi:Stem>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$stems"/>
              <xsl:with-param name="n" select="$transitDayStem + 1"/>
            </xsl:call-template>
          </bazi:Stem>
          <bazi:Branch>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$branches"/>
              <xsl:with-param name="n" select="$transitDayBranch + 1"/>
            </xsl:call-template>
          </bazi:Branch>
        </bazi:DayPillar>
        <bazi:HourPillar>
          <bazi:Stem>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$stems"/>
              <xsl:with-param name="n" select="$transitHourStem + 1"/>
            </xsl:call-template>
          </bazi:Stem>
          <bazi:Branch>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$branches"/>
              <xsl:with-param name="n" select="$transitHourBranch + 1"/>
            </xsl:call-template>
          </bazi:Branch>
        </bazi:HourPillar>
        <bazi:DayMaster>
          <bazi:Stem>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$stems"/>
              <xsl:with-param name="n" select="$transitDMStem + 1"/>
            </xsl:call-template>
          </bazi:Stem>
          <bazi:Element><xsl:value-of select="$transitDMElement"/></bazi:Element>
        </bazi:DayMaster>
      </bazi:TransitPillars>

      <!-- ══════════════════════════════════════════════════════════════════
           PILLAR INTERACTIONS (transit stem × natal stem)
           ════════════════════════════════════════════════════════════════ -->
      <bazi:PillarInteractions>
        <!-- Year pillar interaction -->
        <bazi:YearInteraction>
          <bazi:TenGod>
            <xsl:call-template name="tenGod">
              <xsl:with-param name="dmStem" select="$natalDMStem"/>
              <xsl:with-param name="otherStem" select="$transitYearStem"/>
            </xsl:call-template>
          </bazi:TenGod>
          <bazi:StemClash>
            <xsl:call-template name="stemClash">
              <xsl:with-param name="s1" select="$natalYearStem"/>
              <xsl:with-param name="s2" select="$transitYearStem"/>
            </xsl:call-template>
          </bazi:StemClash>
          <bazi:BranchClash>
            <xsl:call-template name="branchClash">
              <xsl:with-param name="b1" select="$natalYearBranch"/>
              <xsl:with-param name="b2" select="$transitYearBranch"/>
            </xsl:call-template>
          </bazi:BranchClash>
        </bazi:YearInteraction>

        <!-- Month pillar interaction -->
        <bazi:MonthInteraction>
          <bazi:TenGod>
            <xsl:call-template name="tenGod">
              <xsl:with-param name="dmStem" select="$natalDMStem"/>
              <xsl:with-param name="otherStem" select="$transitMonthStem"/>
            </xsl:call-template>
          </bazi:TenGod>
          <bazi:StemClash>
            <xsl:call-template name="stemClash">
              <xsl:with-param name="s1" select="$natalMonthStem"/>
              <xsl:with-param name="s2" select="$transitMonthStem"/>
            </xsl:call-template>
          </bazi:StemClash>
          <bazi:BranchClash>
            <xsl:call-template name="branchClash">
              <xsl:with-param name="b1" select="$natalMonthBranch"/>
              <xsl:with-param name="b2" select="$transitMonthBranch"/>
            </xsl:call-template>
          </bazi:BranchClash>
        </bazi:MonthInteraction>

        <!-- Day pillar interaction -->
        <bazi:DayInteraction>
          <bazi:TenGod>
            <xsl:call-template name="tenGod">
              <xsl:with-param name="dmStem" select="$natalDMStem"/>
              <xsl:with-param name="otherStem" select="$transitDayStem"/>
            </xsl:call-template>
          </bazi:TenGod>
          <bazi:StemClash>
            <xsl:call-template name="stemClash">
              <xsl:with-param name="s1" select="$natalDayStem"/>
              <xsl:with-param name="s2" select="$transitDayStem"/>
            </xsl:call-template>
          </bazi:StemClash>
          <bazi:BranchClash>
            <xsl:call-template name="branchClash">
              <xsl:with-param name="b1" select="$natalDayBranch"/>
              <xsl:with-param name="b2" select="$transitDayBranch"/>
            </xsl:call-template>
          </bazi:BranchClash>
        </bazi:DayInteraction>

        <!-- Hour pillar interaction -->
        <bazi:HourInteraction>
          <bazi:TenGod>
            <xsl:call-template name="tenGod">
              <xsl:with-param name="dmStem" select="$natalDMStem"/>
              <xsl:with-param name="otherStem" select="$transitHourStem"/>
            </xsl:call-template>
          </bazi:TenGod>
          <bazi:StemClash>
            <xsl:call-template name="stemClash">
              <xsl:with-param name="s1" select="$natalHourStem"/>
              <xsl:with-param name="s2" select="$transitHourStem"/>
            </xsl:call-template>
          </bazi:StemClash>
          <bazi:BranchClash>
            <xsl:call-template name="branchClash">
              <xsl:with-param name="b1" select="$natalHourBranch"/>
              <xsl:with-param name="b2" select="$transitHourBranch"/>
            </xsl:call-template>
          </bazi:BranchClash>
        </bazi:HourInteraction>
      </bazi:PillarInteractions>

    </bazi:TransitReport>
  </xsl:template>

  <!-- ══════════════════════════════════════════════════════════════════════
       HELPER TEMPLATES
       ════════════════════════════════════════════════════════════════════ -->

  <!-- Ten God classification: otherStem relative to dmStem -->
  <xsl:template name="tenGod">
    <xsl:param name="dmStem"/>
    <xsl:param name="otherStem"/>
    <xsl:variable name="dmElem" select="$dmStem mod 5"/>
    <xsl:variable name="otherElem" select="$otherStem mod 5"/>
    <xsl:variable name="samePolarity" select="($dmStem mod 2) = ($otherStem mod 2)"/>

    <!-- Element relationship: 0=same, 1=dm produces other, 2=other produces dm, 3=dm controls other, 4=other controls dm -->
    <xsl:variable name="rel" select="($otherElem - $dmElem + 5) mod 5"/>

    <xsl:choose>
      <xsl:when test="$rel = 0">
        <xsl:choose>
          <xsl:when test="$samePolarity">
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$tenGodsSame"/>
              <xsl:with-param name="n" select="1"/>
            </xsl:call-template>
          </xsl:when>
          <xsl:otherwise>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$tenGodsSame"/>
              <xsl:with-param name="n" select="2"/>
            </xsl:call-template>
          </xsl:otherwise>
        </xsl:choose>
      </xsl:when>
      <xsl:when test="$rel = 1">
        <xsl:choose>
          <xsl:when test="$samePolarity">
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$tenGodsProduced"/>
              <xsl:with-param name="n" select="1"/>
            </xsl:call-template>
          </xsl:when>
          <xsl:otherwise>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$tenGodsProduced"/>
              <xsl:with-param name="n" select="2"/>
            </xsl:call-template>
          </xsl:otherwise>
        </xsl:choose>
      </xsl:when>
      <xsl:when test="$rel = 2">
        <xsl:choose>
          <xsl:when test="$samePolarity">
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$tenGodsProduce"/>
              <xsl:with-param name="n" select="1"/>
            </xsl:call-template>
          </xsl:when>
          <xsl:otherwise>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$tenGodsProduce"/>
              <xsl:with-param name="n" select="2"/>
            </xsl:call-template>
          </xsl:otherwise>
        </xsl:choose>
      </xsl:when>
      <xsl:when test="$rel = 3">
        <xsl:choose>
          <xsl:when test="$samePolarity">
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$tenGodsControlled"/>
              <xsl:with-param name="n" select="1"/>
            </xsl:call-template>
          </xsl:when>
          <xsl:otherwise>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$tenGodsControlled"/>
              <xsl:with-param name="n" select="2"/>
            </xsl:call-template>
          </xsl:otherwise>
        </xsl:choose>
      </xsl:when>
      <xsl:otherwise>
        <xsl:choose>
          <xsl:when test="$samePolarity">
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$tenGodsControl"/>
              <xsl:with-param name="n" select="1"/>
            </xsl:call-template>
          </xsl:when>
          <xsl:otherwise>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$tenGodsControl"/>
              <xsl:with-param name="n" select="2"/>
            </xsl:call-template>
          </xsl:otherwise>
        </xsl:choose>
      </xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <!-- Stem clash: stems 0-9, clash pairs are (0,4) (1,5) (2,6) (3,7) (4,8) (5,9) -->
  <xsl:template name="stemClash">
    <xsl:param name="s1"/>
    <xsl:param name="s2"/>
    <xsl:choose>
      <xsl:when test="($s1 + 4) mod 10 = $s2 or ($s2 + 4) mod 10 = $s1">true</xsl:when>
      <xsl:otherwise>false</xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <!-- Branch clash: branches 0-11, clash pairs are (0,6) (1,7) (2,8) (3,9) (4,10) (5,11) -->
  <xsl:template name="branchClash">
    <xsl:param name="b1"/>
    <xsl:param name="b2"/>
    <xsl:choose>
      <xsl:when test="($b1 + 6) mod 12 = $b2 or ($b2 + 6) mod 12 = $b1">true</xsl:when>
      <xsl:otherwise>false</xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <xsl:template name="nthToken">
    <xsl:param name="list"/>
    <xsl:param name="n"/>
    <xsl:param name="delimiter" select="','"/>
    <xsl:choose>
      <xsl:when test="$n = 1">
        <xsl:choose>
          <xsl:when test="contains($list, $delimiter)">
            <xsl:value-of select="substring-before($list, $delimiter)"/>
          </xsl:when>
          <xsl:otherwise>
            <xsl:value-of select="$list"/>
          </xsl:otherwise>
        </xsl:choose>
      </xsl:when>
      <xsl:when test="contains($list, $delimiter)">
        <xsl:call-template name="nthToken">
          <xsl:with-param name="list" select="substring-after($list, $delimiter)"/>
          <xsl:with-param name="n" select="$n - 1"/>
          <xsl:with-param name="delimiter" select="$delimiter"/>
        </xsl:call-template>
      </xsl:when>
    </xsl:choose>
  </xsl:template>

</xsl:stylesheet>
