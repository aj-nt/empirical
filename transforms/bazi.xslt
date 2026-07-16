<?xml version="1.0" encoding="UTF-8"?>
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
  <!-- Format: "stemIdx1,stemIdx2,stemIdx3" — empty string for none -->
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
  <xsl:variable name="tenGodsSame" select="'Bi Jian,Jie Cai'"/>       <!-- same element: same polarity, opposite -->
  <xsl:variable name="tenGodsProduce" select="'Zheng Yin,Pian Yin'"/> <!-- produces DM: same, opposite -->
  <xsl:variable name="tenGodsProduced" select="'Shi Shen,Shang Guan'"/> <!-- DM produces: same, opposite -->
  <xsl:variable name="tenGodsControl" select="'Qi Sha,Zheng Guan'"/>  <!-- controls DM: same, opposite -->
  <xsl:variable name="tenGodsControlled" select="'Pian Cai,Zheng Cai'"/> <!-- DM controls: same, opposite -->

  <!-- ══════════════════════════════════════════════════════════════════════
       ROOT TEMPLATE
       ════════════════════════════════════════════════════════════════════ -->
  <xsl:template match="/BaseChart">
    <bazi:Chart xmlns:bazi="urn:empirical:bazi">
      <bazi:Name><xsl:value-of select="Identity/Name"/></bazi:Name>

      <!-- Compute Four Pillars -->
      <xsl:variable name="year" select="Time/Year"/>
      <xsl:variable name="month" select="Time/Month"/>
      <xsl:variable name="day" select="Time/Day"/>
      <xsl:variable name="hour" select="Time/Hour"/>
      <xsl:variable name="dayJD" select="Time/DayJD"/>

      <!-- Year pillar -->
      <xsl:variable name="yStemIdx" select="(($year - 4) mod 10 + 10) mod 10"/>
      <xsl:variable name="yBranchIdx" select="(($year - 4) mod 12 + 12) mod 12"/>

      <!-- Month pillar: solar term → branch, year stem → first month stem -->
      <xsl:variable name="solarMonth">
        <xsl:call-template name="computeSolarMonth">
          <xsl:with-param name="month" select="$month"/>
          <xsl:with-param name="day" select="$day"/>
        </xsl:call-template>
      </xsl:variable>
      <xsl:variable name="mBranchIdx" select="($solarMonth + 2) mod 12"/>
      <xsl:variable name="firstStemTable" select="'2,4,6,8,0,2,4,6,8,0'"/>
      <xsl:variable name="firstStem">
        <xsl:call-template name="nthToken">
          <xsl:with-param name="list" select="$firstStemTable"/>
          <xsl:with-param name="n" select="$yStemIdx + 1"/>
        </xsl:call-template>
      </xsl:variable>
      <xsl:variable name="mStemIdx" select="($firstStem + $solarMonth) mod 10"/>

      <!-- Day pillar: JD at midnight UTC -->
      <xsl:variable name="dStemIdx" select="($dayJD mod 10 + 10) mod 10"/>
      <xsl:variable name="dBranchIdx" select="(($dayJD + 2) mod 12 + 12) mod 12"/>

      <!-- Hour pillar -->
      <xsl:variable name="hBranchIdx" select="floor(($hour + 1) div 2) mod 12"/>
      <xsl:variable name="hourBaseTable" select="'0,2,4,6,8,0,2,4,6,8'"/>
      <xsl:variable name="hourBase">
        <xsl:call-template name="nthToken">
          <xsl:with-param name="list" select="$hourBaseTable"/>
          <xsl:with-param name="n" select="$dStemIdx + 1"/>
        </xsl:call-template>
      </xsl:variable>
      <xsl:variable name="hStemIdx" select="($hourBase + $hBranchIdx) mod 10"/>

      <!-- Day Master -->
      <xsl:variable name="dmElement">
        <xsl:call-template name="nthToken">
          <xsl:with-param name="list" select="$stemElements"/>
          <xsl:with-param name="n" select="$dStemIdx + 1"/>
        </xsl:call-template>
      </xsl:variable>
      <xsl:variable name="dmYinYang">
        <xsl:call-template name="nthToken">
          <xsl:with-param name="list" select="$stemYinYang"/>
          <xsl:with-param name="n" select="$dStemIdx + 1"/>
        </xsl:call-template>
      </xsl:variable>

      <!-- ══════════════════════════════════════════════════════════════════
           FOUR PILLARS
           ════════════════════════════════════════════════════════════════ -->
      <bazi:Pillars>
        <bazi:Year>
          <bazi:Stem>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$stems"/>
              <xsl:with-param name="n" select="$yStemIdx + 1"/>
            </xsl:call-template>
          </bazi:Stem>
          <bazi:Branch>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$branches"/>
              <xsl:with-param name="n" select="$yBranchIdx + 1"/>
            </xsl:call-template>
          </bazi:Branch>
          <bazi:Animal>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$branchAnimals"/>
              <xsl:with-param name="n" select="$yBranchIdx + 1"/>
            </xsl:call-template>
          </bazi:Animal>
          <bazi:Element>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$stemElements"/>
              <xsl:with-param name="n" select="$yStemIdx + 1"/>
            </xsl:call-template>
          </bazi:Element>
          <bazi:HiddenStems>
            <xsl:call-template name="formatHiddenStems">
              <xsl:with-param name="branchIdx" select="$yBranchIdx"/>
            </xsl:call-template>
          </bazi:HiddenStems>
        </bazi:Year>

        <bazi:Month>
          <bazi:Stem>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$stems"/>
              <xsl:with-param name="n" select="$mStemIdx + 1"/>
            </xsl:call-template>
          </bazi:Stem>
          <bazi:Branch>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$branches"/>
              <xsl:with-param name="n" select="$mBranchIdx + 1"/>
            </xsl:call-template>
          </bazi:Branch>
          <bazi:Animal>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$branchAnimals"/>
              <xsl:with-param name="n" select="$mBranchIdx + 1"/>
            </xsl:call-template>
          </bazi:Animal>
          <bazi:Element>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$stemElements"/>
              <xsl:with-param name="n" select="$mStemIdx + 1"/>
            </xsl:call-template>
          </bazi:Element>
          <bazi:HiddenStems>
            <xsl:call-template name="formatHiddenStems">
              <xsl:with-param name="branchIdx" select="$mBranchIdx"/>
            </xsl:call-template>
          </bazi:HiddenStems>
        </bazi:Month>

        <bazi:Day>
          <bazi:Stem>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$stems"/>
              <xsl:with-param name="n" select="$dStemIdx + 1"/>
            </xsl:call-template>
          </bazi:Stem>
          <bazi:Branch>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$branches"/>
              <xsl:with-param name="n" select="$dBranchIdx + 1"/>
            </xsl:call-template>
          </bazi:Branch>
          <bazi:Animal>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$branchAnimals"/>
              <xsl:with-param name="n" select="$dBranchIdx + 1"/>
            </xsl:call-template>
          </bazi:Animal>
          <bazi:Element>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$stemElements"/>
              <xsl:with-param name="n" select="$dStemIdx + 1"/>
            </xsl:call-template>
          </bazi:Element>
          <bazi:HiddenStems>
            <xsl:call-template name="formatHiddenStems">
              <xsl:with-param name="branchIdx" select="$dBranchIdx"/>
            </xsl:call-template>
          </bazi:HiddenStems>
        </bazi:Day>

        <bazi:Hour>
          <bazi:Stem>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$stems"/>
              <xsl:with-param name="n" select="$hStemIdx + 1"/>
            </xsl:call-template>
          </bazi:Stem>
          <bazi:Branch>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$branches"/>
              <xsl:with-param name="n" select="$hBranchIdx + 1"/>
            </xsl:call-template>
          </bazi:Branch>
          <bazi:Animal>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$branchAnimals"/>
              <xsl:with-param name="n" select="$hBranchIdx + 1"/>
            </xsl:call-template>
          </bazi:Animal>
          <bazi:Element>
            <xsl:call-template name="nthToken">
              <xsl:with-param name="list" select="$stemElements"/>
              <xsl:with-param name="n" select="$hStemIdx + 1"/>
            </xsl:call-template>
          </bazi:Element>
          <bazi:HiddenStems>
            <xsl:call-template name="formatHiddenStems">
              <xsl:with-param name="branchIdx" select="$hBranchIdx"/>
            </xsl:call-template>
          </bazi:HiddenStems>
        </bazi:Hour>
      </bazi:Pillars>

      <!-- ══════════════════════════════════════════════════════════════════
           DAY MASTER
           ════════════════════════════════════════════════════════════════ -->
      <bazi:DayMaster>
        <bazi:Stem>
          <xsl:call-template name="nthToken">
            <xsl:with-param name="list" select="$stems"/>
            <xsl:with-param name="n" select="$dStemIdx + 1"/>
          </xsl:call-template>
        </bazi:Stem>
        <bazi:YinYang><xsl:value-of select="$dmYinYang"/></bazi:YinYang>
        <bazi:Element><xsl:value-of select="$dmElement"/></bazi:Element>
      </bazi:DayMaster>

      <!-- ══════════════════════════════════════════════════════════════════
           TEN GODS (all heavenly stems relative to Day Master)
           ════════════════════════════════════════════════════════════════ -->
      <bazi:TenGods>
        <xsl:call-template name="classifyTenGods">
          <xsl:with-param name="label" select="'Year Stem'"/>
          <xsl:with-param name="stemIdx" select="$yStemIdx"/>
          <xsl:with-param name="dmStemIdx" select="$dStemIdx"/>
        </xsl:call-template>
        <xsl:call-template name="classifyTenGods">
          <xsl:with-param name="label" select="'Month Stem'"/>
          <xsl:with-param name="stemIdx" select="$mStemIdx"/>
          <xsl:with-param name="dmStemIdx" select="$dStemIdx"/>
        </xsl:call-template>
        <xsl:call-template name="classifyTenGods">
          <xsl:with-param name="label" select="'Day Stem'"/>
          <xsl:with-param name="stemIdx" select="$dStemIdx"/>
          <xsl:with-param name="dmStemIdx" select="$dStemIdx"/>
        </xsl:call-template>
        <xsl:call-template name="classifyTenGods">
          <xsl:with-param name="label" select="'Hour Stem'"/>
          <xsl:with-param name="stemIdx" select="$hStemIdx"/>
          <xsl:with-param name="dmStemIdx" select="$dStemIdx"/>
        </xsl:call-template>
        <!-- Hidden stems ten gods -->
        <xsl:call-template name="classifyHiddenTenGods">
          <xsl:with-param name="label" select="'Year Hidden'"/>
          <xsl:with-param name="branchIdx" select="$yBranchIdx"/>
          <xsl:with-param name="dmStemIdx" select="$dStemIdx"/>
        </xsl:call-template>
        <xsl:call-template name="classifyHiddenTenGods">
          <xsl:with-param name="label" select="'Month Hidden'"/>
          <xsl:with-param name="branchIdx" select="$mBranchIdx"/>
          <xsl:with-param name="dmStemIdx" select="$dStemIdx"/>
        </xsl:call-template>
        <xsl:call-template name="classifyHiddenTenGods">
          <xsl:with-param name="label" select="'Day Hidden'"/>
          <xsl:with-param name="branchIdx" select="$dBranchIdx"/>
          <xsl:with-param name="dmStemIdx" select="$dStemIdx"/>
        </xsl:call-template>
        <xsl:call-template name="classifyHiddenTenGods">
          <xsl:with-param name="label" select="'Hour Hidden'"/>
          <xsl:with-param name="branchIdx" select="$hBranchIdx"/>
          <xsl:with-param name="dmStemIdx" select="$dStemIdx"/>
        </xsl:call-template>
      </bazi:TenGods>

      <!-- ══════════════════════════════════════════════════════════════════
           LUCK PILLARS (10 × 10-year periods)
           ════════════════════════════════════════════════════════════════ -->
      <xsl:variable name="direction">
        <xsl:choose>
          <xsl:when test="$dmYinYang = 'Yang'">1</xsl:when>
          <xsl:otherwise>-1</xsl:otherwise>
        </xsl:choose>
      </xsl:variable>

      <bazi:LuckPillars>
        <xsl:call-template name="generateLuckPillars">
          <xsl:with-param name="i" select="1"/>
          <xsl:with-param name="mStemIdx" select="$mStemIdx"/>
          <xsl:with-param name="mBranchIdx" select="$mBranchIdx"/>
          <xsl:with-param name="direction" select="$direction"/>
          <xsl:with-param name="birthYear" select="$year"/>
        </xsl:call-template>
      </bazi:LuckPillars>

    </bazi:Chart>
  </xsl:template>

  <!-- ══════════════════════════════════════════════════════════════════════
       SOLAR MONTH COMPUTATION
       ════════════════════════════════════════════════════════════════════ -->
  <xsl:template name="computeSolarMonth">
    <xsl:param name="month"/>
    <xsl:param name="day"/>
    <!-- Solar terms: (month, day) pairs. Index 0=LiChun(2,4), ..., 11=XiaoHan(1,6) -->
    <xsl:choose>
      <!-- Before Li Chun (Feb 4) → solar month 11 -->
      <xsl:when test="$month &lt; 2 or ($month = 2 and $day &lt; 4)">11</xsl:when>
      <!-- Before Jing Zhe (Mar 6) → solar month 0 -->
      <xsl:when test="$month &lt; 3 or ($month = 3 and $day &lt; 6)">0</xsl:when>
      <!-- Before Qing Ming (Apr 5) → solar month 1 -->
      <xsl:when test="$month &lt; 4 or ($month = 4 and $day &lt; 5)">1</xsl:when>
      <!-- Before Li Xia (May 6) → solar month 2 -->
      <xsl:when test="$month &lt; 5 or ($month = 5 and $day &lt; 6)">2</xsl:when>
      <!-- Before Mang Zhong (Jun 6) → solar month 3 -->
      <xsl:when test="$month &lt; 6 or ($month = 6 and $day &lt; 6)">3</xsl:when>
      <!-- Before Xiao Shu (Jul 7) → solar month 4 -->
      <xsl:when test="$month &lt; 7 or ($month = 7 and $day &lt; 7)">4</xsl:when>
      <!-- Before Li Qiu (Aug 8) → solar month 5 -->
      <xsl:when test="$month &lt; 8 or ($month = 8 and $day &lt; 8)">5</xsl:when>
      <!-- Before Bai Lu (Sep 8) → solar month 6 -->
      <xsl:when test="$month &lt; 9 or ($month = 9 and $day &lt; 8)">6</xsl:when>
      <!-- Before Han Lu (Oct 8) → solar month 7 -->
      <xsl:when test="$month &lt; 10 or ($month = 10 and $day &lt; 8)">7</xsl:when>
      <!-- Before Li Dong (Nov 7) → solar month 8 -->
      <xsl:when test="$month &lt; 11 or ($month = 11 and $day &lt; 7)">8</xsl:when>
      <!-- Before Da Xue (Dec 7) → solar month 9 -->
      <xsl:when test="$month &lt; 12 or ($month = 12 and $day &lt; 7)">9</xsl:when>
      <!-- Before Xiao Han (Jan 6) → solar month 10 -->
      <xsl:when test="$month = 12 and $day &lt; 7">9</xsl:when>
      <xsl:when test="$month = 1 and $day &lt; 6">10</xsl:when>
      <!-- After Xiao Han → solar month 11 -->
      <xsl:otherwise>11</xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <!-- ══════════════════════════════════════════════════════════════════════
       HIDDEN STEMS FORMATTING
       ════════════════════════════════════════════════════════════════════ -->
  <xsl:template name="formatHiddenStems">
    <xsl:param name="branchIdx"/>
    <xsl:variable name="hidden">
      <xsl:choose>
        <xsl:when test="$branchIdx = 0"><xsl:value-of select="$hiddenStems0"/></xsl:when>
        <xsl:when test="$branchIdx = 1"><xsl:value-of select="$hiddenStems1"/></xsl:when>
        <xsl:when test="$branchIdx = 2"><xsl:value-of select="$hiddenStems2"/></xsl:when>
        <xsl:when test="$branchIdx = 3"><xsl:value-of select="$hiddenStems3"/></xsl:when>
        <xsl:when test="$branchIdx = 4"><xsl:value-of select="$hiddenStems4"/></xsl:when>
        <xsl:when test="$branchIdx = 5"><xsl:value-of select="$hiddenStems5"/></xsl:when>
        <xsl:when test="$branchIdx = 6"><xsl:value-of select="$hiddenStems6"/></xsl:when>
        <xsl:when test="$branchIdx = 7"><xsl:value-of select="$hiddenStems7"/></xsl:when>
        <xsl:when test="$branchIdx = 8"><xsl:value-of select="$hiddenStems8"/></xsl:when>
        <xsl:when test="$branchIdx = 9"><xsl:value-of select="$hiddenStems9"/></xsl:when>
        <xsl:when test="$branchIdx = 10"><xsl:value-of select="$hiddenStems10"/></xsl:when>
        <xsl:when test="$branchIdx = 11"><xsl:value-of select="$hiddenStems11"/></xsl:when>
      </xsl:choose>
    </xsl:variable>
    <xsl:call-template name="formatStemList">
      <xsl:with-param name="list" select="$hidden"/>
    </xsl:call-template>
  </xsl:template>

  <xsl:template name="formatStemList">
    <xsl:param name="list"/>
    <xsl:if test="$list != ''">
      <xsl:variable name="first" select="substring-before(concat($list, ','), ',')"/>
      <xsl:variable name="rest" select="substring-after($list, ',')"/>
      <bazi:HiddenStem>
        <xsl:call-template name="nthToken">
          <xsl:with-param name="list" select="$stems"/>
          <xsl:with-param name="n" select="number($first) + 1"/>
        </xsl:call-template>
      </bazi:HiddenStem>
      <xsl:if test="$rest != ''">
        <xsl:call-template name="formatStemList">
          <xsl:with-param name="list" select="$rest"/>
        </xsl:call-template>
      </xsl:if>
    </xsl:if>
  </xsl:template>

  <!-- ══════════════════════════════════════════════════════════════════════
       TEN GODS CLASSIFICATION
       ════════════════════════════════════════════════════════════════════ -->

  <!-- Element index: Wood=0, Fire=1, Earth=2, Metal=3, Water=4 -->
  <xsl:template name="elementIndex">
    <xsl:param name="element"/>
    <xsl:choose>
      <xsl:when test="$element = 'Wood'">0</xsl:when>
      <xsl:when test="$element = 'Fire'">1</xsl:when>
      <xsl:when test="$element = 'Earth'">2</xsl:when>
      <xsl:when test="$element = 'Metal'">3</xsl:when>
      <xsl:when test="$element = 'Water'">4</xsl:when>
      <xsl:otherwise>-1</xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <!-- Classify a single stem relative to Day Master -->
  <xsl:template name="classifyTenGods">
    <xsl:param name="label"/>
    <xsl:param name="stemIdx"/>
    <xsl:param name="dmStemIdx"/>

    <xsl:variable name="stemEl">
      <xsl:call-template name="nthToken">
        <xsl:with-param name="list" select="$stemElements"/>
        <xsl:with-param name="n" select="$stemIdx + 1"/>
      </xsl:call-template>
    </xsl:variable>
    <xsl:variable name="stemYY">
      <xsl:call-template name="nthToken">
        <xsl:with-param name="list" select="$stemYinYang"/>
        <xsl:with-param name="n" select="$stemIdx + 1"/>
      </xsl:call-template>
    </xsl:variable>
    <xsl:variable name="dmEl">
      <xsl:call-template name="nthToken">
        <xsl:with-param name="list" select="$stemElements"/>
        <xsl:with-param name="n" select="$dmStemIdx + 1"/>
      </xsl:call-template>
    </xsl:variable>
    <xsl:variable name="dmYY">
      <xsl:call-template name="nthToken">
        <xsl:with-param name="list" select="$stemYinYang"/>
        <xsl:with-param name="n" select="$dmStemIdx + 1"/>
      </xsl:call-template>
    </xsl:variable>

    <xsl:variable name="sElIdx">
      <xsl:call-template name="elementIndex">
        <xsl:with-param name="element" select="$stemEl"/>
      </xsl:call-template>
    </xsl:variable>
    <xsl:variable name="dElIdx">
      <xsl:call-template name="elementIndex">
        <xsl:with-param name="element" select="$dmEl"/>
      </xsl:call-template>
    </xsl:variable>

    <xsl:variable name="samePolarity" select="$stemYY = $dmYY"/>
    <xsl:variable name="polarityIdx">
      <xsl:choose>
        <xsl:when test="$samePolarity">0</xsl:when>
        <xsl:otherwise>1</xsl:otherwise>
      </xsl:choose>
    </xsl:variable>

    <!-- Determine relationship category -->
    <xsl:variable name="category">
      <xsl:choose>
        <!-- Same element -->
        <xsl:when test="$sElIdx = $dElIdx">same</xsl:when>
        <!-- Producing DM: (dElIdx + 4) mod 5 -->
        <xsl:when test="$sElIdx = ($dElIdx + 4) mod 5">produce</xsl:when>
        <!-- DM produces: (dElIdx + 1) mod 5 -->
        <xsl:when test="$sElIdx = ($dElIdx + 1) mod 5">produced</xsl:when>
        <!-- Controls DM: (dElIdx + 3) mod 5 -->
        <xsl:when test="$sElIdx = ($dElIdx + 3) mod 5">control</xsl:when>
        <!-- DM controls: (dElIdx + 2) mod 5 -->
        <xsl:otherwise>controlled</xsl:otherwise>
      </xsl:choose>
    </xsl:variable>

    <xsl:variable name="godList">
      <xsl:choose>
        <xsl:when test="$category = 'same'"><xsl:value-of select="$tenGodsSame"/></xsl:when>
        <xsl:when test="$category = 'produce'"><xsl:value-of select="$tenGodsProduce"/></xsl:when>
        <xsl:when test="$category = 'produced'"><xsl:value-of select="$tenGodsProduced"/></xsl:when>
        <xsl:when test="$category = 'control'"><xsl:value-of select="$tenGodsControl"/></xsl:when>
        <xsl:otherwise><xsl:value-of select="$tenGodsControlled"/></xsl:otherwise>
      </xsl:choose>
    </xsl:variable>

    <bazi:TenGod>
      <bazi:Source><xsl:value-of select="$label"/></bazi:Source>
      <bazi:Stem>
        <xsl:call-template name="nthToken">
          <xsl:with-param name="list" select="$stems"/>
          <xsl:with-param name="n" select="$stemIdx + 1"/>
        </xsl:call-template>
      </bazi:Stem>
      <bazi:Element><xsl:value-of select="$stemEl"/></bazi:Element>
      <bazi:God>
        <xsl:call-template name="nthToken">
          <xsl:with-param name="list" select="$godList"/>
          <xsl:with-param name="n" select="$polarityIdx + 1"/>
        </xsl:call-template>
      </bazi:God>
    </bazi:TenGod>
  </xsl:template>

  <!-- Classify hidden stems for a branch -->
  <xsl:template name="classifyHiddenTenGods">
    <xsl:param name="label"/>
    <xsl:param name="branchIdx"/>
    <xsl:param name="dmStemIdx"/>

    <xsl:variable name="hidden">
      <xsl:choose>
        <xsl:when test="$branchIdx = 0"><xsl:value-of select="$hiddenStems0"/></xsl:when>
        <xsl:when test="$branchIdx = 1"><xsl:value-of select="$hiddenStems1"/></xsl:when>
        <xsl:when test="$branchIdx = 2"><xsl:value-of select="$hiddenStems2"/></xsl:when>
        <xsl:when test="$branchIdx = 3"><xsl:value-of select="$hiddenStems3"/></xsl:when>
        <xsl:when test="$branchIdx = 4"><xsl:value-of select="$hiddenStems4"/></xsl:when>
        <xsl:when test="$branchIdx = 5"><xsl:value-of select="$hiddenStems5"/></xsl:when>
        <xsl:when test="$branchIdx = 6"><xsl:value-of select="$hiddenStems6"/></xsl:when>
        <xsl:when test="$branchIdx = 7"><xsl:value-of select="$hiddenStems7"/></xsl:when>
        <xsl:when test="$branchIdx = 8"><xsl:value-of select="$hiddenStems8"/></xsl:when>
        <xsl:when test="$branchIdx = 9"><xsl:value-of select="$hiddenStems9"/></xsl:when>
        <xsl:when test="$branchIdx = 10"><xsl:value-of select="$hiddenStems10"/></xsl:when>
        <xsl:when test="$branchIdx = 11"><xsl:value-of select="$hiddenStems11"/></xsl:when>
      </xsl:choose>
    </xsl:variable>

    <xsl:call-template name="classifyHiddenRecurse">
      <xsl:with-param name="label" select="$label"/>
      <xsl:with-param name="list" select="$hidden"/>
      <xsl:with-param name="dmStemIdx" select="$dmStemIdx"/>
    </xsl:call-template>
  </xsl:template>

  <xsl:template name="classifyHiddenRecurse">
    <xsl:param name="label"/>
    <xsl:param name="list"/>
    <xsl:param name="dmStemIdx"/>

    <xsl:if test="$list != ''">
      <xsl:variable name="first" select="substring-before(concat($list, ','), ',')"/>
      <xsl:variable name="rest" select="substring-after($list, ',')"/>
      <xsl:call-template name="classifyTenGods">
        <xsl:with-param name="label" select="$label"/>
        <xsl:with-param name="stemIdx" select="number($first)"/>
        <xsl:with-param name="dmStemIdx" select="$dmStemIdx"/>
      </xsl:call-template>
      <xsl:if test="$rest != ''">
        <xsl:call-template name="classifyHiddenRecurse">
          <xsl:with-param name="label" select="$label"/>
          <xsl:with-param name="list" select="$rest"/>
          <xsl:with-param name="dmStemIdx" select="$dmStemIdx"/>
        </xsl:call-template>
      </xsl:if>
    </xsl:if>
  </xsl:template>

  <!-- ══════════════════════════════════════════════════════════════════════
       LUCK PILLARS GENERATION
       ════════════════════════════════════════════════════════════════════ -->
  <xsl:template name="generateLuckPillars">
    <xsl:param name="i"/>         <!-- 1-based index -->
    <xsl:param name="mStemIdx"/>
    <xsl:param name="mBranchIdx"/>
    <xsl:param name="direction"/>
    <xsl:param name="birthYear"/>

    <xsl:if test="$i &lt;= 10">
      <xsl:variable name="lpStemIdx" select="(($mStemIdx + $direction * $i) mod 10 + 10) mod 10"/>
      <xsl:variable name="lpBranchIdx" select="(($mBranchIdx + $direction * $i) mod 12 + 12) mod 12"/>
      <xsl:variable name="startYear" select="$birthYear + ($i - 1) * 10"/>
      <xsl:variable name="endYear" select="$birthYear + $i * 10"/>

      <bazi:LuckPillar>
        <bazi:Index><xsl:value-of select="$i"/></bazi:Index>
        <bazi:Stem>
          <xsl:call-template name="nthToken">
            <xsl:with-param name="list" select="$stems"/>
            <xsl:with-param name="n" select="$lpStemIdx + 1"/>
          </xsl:call-template>
        </bazi:Stem>
        <bazi:Branch>
          <xsl:call-template name="nthToken">
            <xsl:with-param name="list" select="$branches"/>
            <xsl:with-param name="n" select="$lpBranchIdx + 1"/>
          </xsl:call-template>
        </bazi:Branch>
        <bazi:Animal>
          <xsl:call-template name="nthToken">
            <xsl:with-param name="list" select="$branchAnimals"/>
            <xsl:with-param name="n" select="$lpBranchIdx + 1"/>
          </xsl:call-template>
        </bazi:Animal>
        <bazi:Element>
          <xsl:call-template name="nthToken">
            <xsl:with-param name="list" select="$stemElements"/>
            <xsl:with-param name="n" select="$lpStemIdx + 1"/>
          </xsl:call-template>
        </bazi:Element>
        <bazi:StartYear><xsl:value-of select="$startYear"/></bazi:StartYear>
        <bazi:EndYear><xsl:value-of select="$endYear"/></bazi:EndYear>
      </bazi:LuckPillar>

      <xsl:call-template name="generateLuckPillars">
        <xsl:with-param name="i" select="$i + 1"/>
        <xsl:with-param name="mStemIdx" select="$mStemIdx"/>
        <xsl:with-param name="mBranchIdx" select="$mBranchIdx"/>
        <xsl:with-param name="direction" select="$direction"/>
        <xsl:with-param name="birthYear" select="$birthYear"/>
      </xsl:call-template>
    </xsl:if>
  </xsl:template>

  <!-- ══════════════════════════════════════════════════════════════════════
       CORE HELPERS
       ════════════════════════════════════════════════════════════════════ -->

  <!-- nth comma-separated token (1-indexed) -->
  <xsl:template name="nthToken">
    <xsl:param name="list"/>
    <xsl:param name="n"/>
    <xsl:choose>
      <xsl:when test="$n = 1">
        <xsl:value-of select="substring-before(concat($list, ','), ',')"/>
      </xsl:when>
      <xsl:otherwise>
        <xsl:call-template name="nthToken">
          <xsl:with-param name="list" select="substring-after($list, ',')"/>
          <xsl:with-param name="n" select="$n - 1"/>
        </xsl:call-template>
      </xsl:otherwise>
    </xsl:choose>
  </xsl:template>

</xsl:stylesheet>
